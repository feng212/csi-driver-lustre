package lustre_driver

import (
	"context"
	"csi-driver-lustre/pkg/lustre-driver/lustrefs"
	"fmt"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"
)

var (
	// volumeCaps represents how the volume could be accessed.
	volumeCaps = []csi.VolumeCapability_AccessMode{
		{
			Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
		},
		{
			Mode: csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER,
		},
	}

	// controllerCaps represents the capability of controller service
	controllerCaps = []csi.ControllerServiceCapability_RPC_Type{
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
		csi.ControllerServiceCapability_RPC_EXPAND_VOLUME,
	}
)

// controllerService represents the controller service of CSI driver
type controllerService struct {
	inFlight *lustrefs.InFlight
	lustre   *lustrefs.Lustre
}

func NewControllerService() {

}

func (cs *controllerService) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	klog.V(4).InfoS("CreateVolume: called", "args", *req)
	volName := req.GetName()
	if len(volName) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume name not provided")
	}
	volCaps := req.GetVolumeCapabilities()
	if isValidVolumeCapabilities(volCaps) != nil {
		return nil, status.Error(codes.InvalidArgument, "Volume capabilities not supported")
	}
	// check if a request is already in-flight
	//if ok := cs.inFlight.Insert(volName); !ok {
	//	msg := fmt.Sprintf("Create volume request for %s is already in progress", volName)
	//	return nil, status.Error(codes.Aborted, msg)
	//}
	//defer cs.inFlight.Delete(volName)
	// create a new volume with idempotency
	volParam := req.GetParameters()
	if volParam == nil {
		volParam = make(map[string]string)
	}
	cs.lustre.StorageType = paramStorageType
	if val, ok := volParam[paramServer]; ok {
		cs.lustre.MountName = val
	}
	if val, ok := volParam[paramSubDir]; ok {
		cs.lustre.MountPoint = val
	}
	if val, ok := volParam[paramDIRPid]; ok {
		cs.lustre.ProjectId = val
	}
	if val, ok := volParam[paramDIRUid]; ok {
		cs.lustre.Uid = val
	}
	err := cs.lustre.LustreMount()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to mount nfs server: %v", err.Error())
	}

	return nil, nil
}
func (cs *controllerService) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	return nil, nil
}
func (cs *controllerService) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	return nil, nil
}
func (cs *controllerService) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	return nil, nil
}
func (cs *controllerService) ValidateVolumeCapabilities(ctx context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	return nil, nil
}
func (cs *controllerService) ListVolumes(ctx context.Context, req *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	return nil, nil
}
func (cs *controllerService) GetCapacity(ctx context.Context, req *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	return nil, nil
}
func (cs *controllerService) ControllerGetCapabilities(ctx context.Context, req *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	return nil, nil
}
func (cs *controllerService) CreateSnapshot(ctx context.Context, req *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	return nil, nil
}
func (cs *controllerService) DeleteSnapshot(ctx context.Context, req *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	return nil, nil
}
func (cs *controllerService) ListSnapshots(ctx context.Context, req *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	return nil, nil
}
func (cs *controllerService) ControllerExpandVolume(ctx context.Context, req *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	return nil, nil
}
func (cs *controllerService) ControllerGetVolume(ctx context.Context, req *csi.ControllerGetVolumeRequest) (*csi.ControllerGetVolumeResponse, error) {
	return nil, nil
}
func (cs *controllerService) ControllerModifyVolume(ctx context.Context, req *csi.ControllerModifyVolumeRequest) (*csi.ControllerModifyVolumeResponse, error) {
	return nil, nil
}
func isValidVolumeCapabilities(caps []*csi.VolumeCapability) error {
	if len(caps) == 0 {
		return fmt.Errorf("volume capabilities missing in request")
	}
	hasSupport := func(cap *csi.VolumeCapability) bool {
		for _, c := range volumeCaps {
			if c.GetMode() == cap.AccessMode.GetMode() {
				return true
			}
		}
		return false
	}
	for _, c := range caps {
		if c.GetBlock() != nil {
			return fmt.Errorf("block volume capability not supported")
		}
		if !hasSupport(c) {
			return fmt.Errorf("mode not supported")
		}
	}
	return nil
}
