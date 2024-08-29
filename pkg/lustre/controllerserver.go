package lustre

import (
	"context"
	"csi-driver-lustre/pkg/lustre-driver/lustrefs"
	"fmt"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"
)

const (
	GiB                     = 1024 * 1024 * 1024
	volumeContextFsTYPE     = "fstype"
	volumeContextServerName = "servername"
	volumeContextMountName  = "mountname"
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

type ControllerServer struct {
	csi.UnimplementedControllerServer
	Driver *Driver
}

func (cs *ControllerServer) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
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
	if ok := cs.Driver.volumeLocks.Insert(volName); !ok {
		msg := fmt.Sprintf("Create volume request for %s is already in progress", volName)
		return nil, status.Error(codes.Aborted, msg)
	}
	reqCapacity := req.GetCapacityRange()
	if reqCapacity == nil {
		cs.lustre.CapacityGiB = lustrefs.DefaultVolumeSize

	}
	// create a new volume with idempotency
	volParam := req.GetParameters()
	if volParam == nil {
		volParam = make(map[string]string)
	}
	cs.lustre.StorageType = paramFsType
	if val, ok := volParam[paramServer]; ok {
		cs.lustre.ServerName = val
	}
	if val, ok := volParam[paramBaseDir]; ok {
		cs.lustre.MountPoint = val
	}
	if val, ok := volParam[paramSubDir]; ok {
		cs.lustre.SubDir = val
	}
	if val, ok := volParam[paramDIRPid]; ok {
		cs.lustre.ProjectId = val
	}
	if val, ok := volParam[paramDIRUid]; ok {
		cs.lustre.Uid = val
	}
	err := cs.lustre.CreateFs()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to mount nfs server: %v", err.Error())
	}

	return newCreateVolumeResponse(cs.lustre), nil
}
	return nil, nil
}
func (cs *ControllerServer) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	return nil, nil
}
func (cs *ControllerServer) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	return nil, nil
}
func (cs *ControllerServer) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	return nil, nil
}
func (cs *ControllerServer) ValidateVolumeCapabilities(ctx context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	return nil, nil
}
func (cs *ControllerServer) ListVolumes(ctx context.Context, req *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	return nil, nil
}
func (cs *ControllerServer) GetCapacity(ctx context.Context, req *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	return nil, nil
}
func (cs *ControllerServer) ControllerGetCapabilities(ctx context.Context, req *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	return nil, nil
}
func (cs *ControllerServer) CreateSnapshot(ctx context.Context, req *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	return nil, nil
}
func (cs *ControllerServer) DeleteSnapshot(ctx context.Context, req *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	return nil, nil
}
func (cs *ControllerServer) ListSnapshots(ctx context.Context, req *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	return nil, nil
}
func (cs *ControllerServer) ControllerExpandVolume(ctx context.Context, req *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	return nil, nil
}
func (cs *ControllerServer) ControllerGetVolume(ctx context.Context, req *csi.ControllerGetVolumeRequest) (*csi.ControllerGetVolumeResponse, error) {
	return nil, nil
}
func (cs *ControllerServer) ControllerModifyVolume(ctx context.Context, req *csi.ControllerModifyVolumeRequest) (*csi.ControllerModifyVolumeResponse, error) {
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
