package lustre_driver

import (
	"context"
	"csi-driver-lustre/pkg/lustre-driver/lustrefs"
	"fmt"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"
	"os"
)

type nodeService struct {
	inFlight *lustrefs.InFlight
}

const VolumeOperationAlreadyExists = "An operation with the given volume=%q and target=%q is already in progress"

func (ns *nodeService) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	return nil, nil
}
func (ns *nodeService) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	return nil, nil
}
func (ns *nodeService) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	klog.V(4).InfoS("NodePublishVolume: called with", "args", *req)
	volumeID := req.GetVolumeId()
	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID not provided")
	}
	context := req.GetVolumeContext()
	server := context[volumeContextServerName]
	if len(server) == 0 {
		return nil, status.Error(codes.InvalidArgument, "server is not provided")
	}
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
	rpcKey := fmt.Sprintf("%s-%s", volumeID, targetPath)
	if ok := ns.inFlight.Insert(rpcKey); !ok {
		return nil, status.Errorf(codes.Aborted, VolumeOperationAlreadyExists, volumeID, targetPath)
	}
	defer func() {
		klog.V(4).InfoS("NodePublishVolume: volume operation finished", "rpcKey", rpcKey)
		ns.inFlight.Delete(rpcKey)
	}()

	//mounted, err := ns.isMounted(source, target)
	if err := d.mounter.Mount(source, target, "lustre", mountOptions); err != nil {
		os.Remove(target)
		return nil, status.Errorf(codes.Internal, "Could not mount %q at %q: %v", source, target, err)
	}
	klog.V(5).InfoS("NodePublishVolume: was mounted", "target", target)
	return nil, nil
}
func (ns *nodeService) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	return nil, nil
}
func (ns *nodeService) NodeGetVolumeStats(ctx context.Context, req *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	return nil, nil
}
func (ns *nodeService) NodeExpandVolume(ctx context.Context, req *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	return nil, nil
}
func (ns *nodeService) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	return nil, nil
}
func (ns *nodeService) NodeGetInfo(ctx context.Context, req *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	return nil, nil
}
