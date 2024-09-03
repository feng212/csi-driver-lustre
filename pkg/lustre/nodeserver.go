package lustre

import (
	"context"
	"fmt"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"
	"k8s.io/mount-utils"
	"os"
	"strings"
)

const VolumeOperationAlreadyExists = "An operation with the given volume=%q and target=%q is already in progress"

type NodeServer struct {
	csi.UnimplementedNodeServer
	Driver *Driver
}

func (ns *NodeServer) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {

	return nil, nil
}
func (ns *NodeServer) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	return nil, nil
}
func (ns *NodeServer) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	klog.V(4).InfoS("NodePublishVolume: called with", "args", *req)
	volumeID := req.GetVolumeId()
	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID not provided")
	}
	context := req.GetVolumeContext()
	server := context["server"]
	baseDir := context["base_dir"]
	subDir := context["subdir"]
	if len(server) == 0 {
		return nil, status.Error(codes.InvalidArgument, "server is not provided")
	}
	if len(baseDir) == 0 {
		return nil, status.Error(codes.InvalidArgument, "subdir is not provided")
	}
	source := fmt.Sprintf("%s/%s", server, subDir)
	targetPath := req.GetTargetPath()
	if len(targetPath) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Target path not provided")
	}
	volCap := req.GetVolumeCapability()
	if volCap == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume capability not provided")
	}
	//if !isValidVolumeCapabilities([]*csi.VolumeCapability{volCap}) {
	//	return nil, status.Error(codes.InvalidArgument, "Volume capability not supported")
	//}
	//rpcKey := fmt.Sprintf("%s-%s", volumeID, targetPath)
	//if ok := ns.Driver.volumeLocks.Insert(rpcKey); !ok {
	//	return nil, status.Errorf(codes.Aborted, VolumeOperationAlreadyExists, volumeID, targetPath)
	//}
	//defer func() {
	//	klog.V(4).InfoS("NodePublishVolume: volume operation finished", "rpcKey", rpcKey)
	//	ns.Driver.volumeLocks.Delete(rpcKey)
	//}()
	mountOptions := volCap.GetMount().GetMountFlags()
	if req.GetReadonly() {
		mountOptions = append(mountOptions, "ro")
	}

	//mounted, err := ns.isMounted(source, target)
	lustre := &Lustre{Mount: mount.New("")}
	err := lustre.Mount.Mount(source, targetPath, "lustre", mountOptions)
	if err != nil {
		if os.IsPermission(err) {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}
		if strings.Contains(err.Error(), "invalid argument") {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return nil, nil
}
func (ns *NodeServer) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	return nil, nil
}
func (ns *NodeServer) NodeGetVolumeStats(ctx context.Context, req *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	return nil, nil
}
func (ns *NodeServer) NodeExpandVolume(ctx context.Context, req *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	return nil, nil
}
func (ns *NodeServer) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	return nil, nil
}
func (ns *NodeServer) NodeGetInfo(ctx context.Context, req *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	return nil, nil
}
