package lustre

import (
	"context"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"
	"k8s.io/mount-utils"
	"os"
)

const VolumeOperationAlreadyExists = "An operation with the given volume=%q and target=%q is already in progress"

type NodeServer struct {
	csi.UnimplementedNodeServer
	Driver *Driver
	Mount  mount.Interface
}

// NodeStageVolume prepares the volume to be published. For Lustre, this might not require any special staging.
func (ns *NodeServer) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	klog.V(4).InfoS("NodeStageVolume called", "volumeId", req.GetVolumeId())
	// No special staging needed for Lustre, just return success
	return &csi.NodeStageVolumeResponse{}, nil
}

// NodeUnstageVolume removes the staged volume. This can be used to clean up staged resources.
func (ns *NodeServer) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	klog.V(4).InfoS("NodeUnstageVolume called", "volumeId", req.GetVolumeId())
	return &csi.NodeUnstageVolumeResponse{}, nil
}

// NodePublishVolume mounts the Lustre volume to the target path on the node.
func (ns *NodeServer) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	klog.V(4).InfoS("NodePublishVolume called", "volumeId", req.GetVolumeId(), "targetPath", req.GetTargetPath())

	// 校验请求参数
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID not provided")
	}
	if len(req.GetTargetPath()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Target path not provided")
	}
	if req.GetVolumeCapability() == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume capability not provided")
	}

	// 处理 ReadOnly 的情况
	readOnly := req.GetReadonly()

	// 获取卷的上下文，比如 Lustre 文件系统需要的 servername 和 mountname
	volumeContext := req.GetVolumeContext()
	serverName, ok := volumeContext["servername"]
	if !ok || len(serverName) == 0 {
		return nil, status.Error(codes.InvalidArgument, "servername not provided in volume context")
	}

	targetPath := req.GetTargetPath()

	// 检查目标路径是否已经挂载
	notMnt, err := ns.Mount.IsLikelyNotMountPoint(targetPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, status.Errorf(codes.Internal, "could not determine if %s is a mount point: %v", targetPath, err)
	}
	if !notMnt {
		klog.V(4).InfoS("Volume is already mounted", "targetPath", targetPath)
		return &csi.NodePublishVolumeResponse{}, nil
	}

	// 创建目标路径，如果它不存在
	if err := os.MkdirAll(targetPath, 0750); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create target path %s: %v", targetPath, err)
	}

	// 挂载选项
	mountOptions := []string{}
	if readOnly {
		mountOptions = append(mountOptions, "ro")
	}

	// 执行挂载操作
	klog.V(4).InfoS("Mounting volume", "source", serverName, "targetPath", targetPath, "options", mountOptions)
	err = ns.Mount.Mount(serverName, targetPath, "lustre", mountOptions)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to mount %s at %s: %v", serverName, targetPath, err)
	}

	klog.V(4).InfoS("NodePublishVolume successful", "volumeId", req.GetVolumeId(), "targetPath", targetPath)
	return &csi.NodePublishVolumeResponse{}, nil
}

// NodeUnpublishVolume unmounts the Lustre volume from the target path.
func (ns *NodeServer) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	klog.V(4).InfoS("NodeUnpublishVolume called", "volumeId", req.GetVolumeId(), "targetPath", req.GetTargetPath())

	targetPath := req.GetTargetPath()

	// 检查目标路径是否已经挂载
	notMnt, err := ns.Mount.IsLikelyNotMountPoint(targetPath)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to check if targetPath %s is a mount point: %v", targetPath, err)
	}
	if notMnt {
		klog.V(4).InfoS("Volume is not mounted", "targetPath", targetPath)
		return &csi.NodeUnpublishVolumeResponse{}, nil
	}

	// 卸载卷
	if err := ns.Mount.Unmount(targetPath); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to unmount targetPath %s: %v", targetPath, err)
	}

	klog.V(4).InfoS("NodeUnpublishVolume successful", "volumeId", req.GetVolumeId(), "targetPath", targetPath)
	return &csi.NodeUnpublishVolumeResponse{}, nil
}

// NodeGetVolumeStats returns volume usage statistics (not typically applicable for Lustre).
func (ns *NodeServer) NodeGetVolumeStats(ctx context.Context, req *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "NodeGetVolumeStats is not implemented")
}

// NodeExpandVolume expands the volume if possible (Lustre doesn't typically support expansion at the node level).
func (ns *NodeServer) NodeExpandVolume(ctx context.Context, req *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "NodeExpandVolume is not implemented")
}

// NodeGetCapabilities returns the supported capabilities of the node.
func (ns *NodeServer) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	capabilities := []*csi.NodeServiceCapability{
		{
			Type: &csi.NodeServiceCapability_Rpc{
				Rpc: &csi.NodeServiceCapability_RPC{
					Type: csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME,
				},
			},
		},
	}
	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: capabilities,
	}, nil
}

// NodeGetInfo returns node-specific information.
func (ns *NodeServer) NodeGetInfo(ctx context.Context, req *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	return &csi.NodeGetInfoResponse{
		NodeId: ns.Driver.NodeId,
	}, nil
}
