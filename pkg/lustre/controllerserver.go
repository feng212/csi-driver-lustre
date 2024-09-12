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

const (
	DefaultVolumeSize       int64 = 1200
	volumeContextFsTYPE           = "fstype"
	volumeContextServerName       = "servername"
	volumeContextMountName        = "mountname"
	volumeContextSubName          = "subname"
	idServer                      = iota
	idBaseDir
	idSubDir
	idUUID
	idOnDelete
	totalIDElements // Always last
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
	klog.V(4).InfoS("CreateVolume: called", "volumeName", req.GetName(), "args", *req)

	volName := req.GetName()
	if len(volName) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume Name not provided")
	}

	volCaps := req.GetVolumeCapabilities()
	if err := isValidVolumeCapabilities(volCaps); err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid volume capabilities: "+err.Error())
	}

	// 设置容量，如果未指定，则使用默认值
	reqCapacity := req.GetCapacityRange().GetRequiredBytes()
	if reqCapacity == 0 {
		reqCapacity = DefaultVolumeSize
	}

	if ok := cs.Driver.VolumeLocks.Insert(volName); !ok {
		msg := fmt.Sprintf("Create volume request for %s is already in progress", volName)
		return nil, status.Error(codes.Aborted, msg)
	}

	lustre := &Lustre{
		Mount:       mount.New(""),
		OnDelete:    cs.Driver.DefaultOnDeletePolicy,
		StorageType: paramFsType,
	}

	volParam := req.GetParameters()
	if volParam == nil {
		volParam = make(map[string]string)
	}

	// 设置 Lustre 参数
	cs.setLustreParameters(volParam, lustre)
	if lustre.SubDir == "" {
		lustre.SubDir = req.GetName()
	}

	// 校验 OnDelete 参数值
	if err := validateOnDeleteValue(lustre.OnDelete); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	lustre.FSId = getVolumeIDFromLustreVol(lustre)

	// 挂载操作
	if err := cs.internalMount(ctx, lustre); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to mount lustre: %v", err)
	}

	internalVolumePath := getInternalMountPath(lustre)
	if err := os.MkdirAll(internalVolumePath, 0777); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to make subdirectory: %v", err)
	}

	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      lustre.FSId,
			CapacityBytes: reqCapacity, // 设置容量
			VolumeContext: volParam,
			ContentSource: req.GetVolumeContentSource(),
		},
	}, nil
}

// setLustreParameters 函数，用于提取并设置参数
func (cs *ControllerServer) setLustreParameters(volParam map[string]string, lustre *Lustre) {
	if val, ok := volParam[paramServer]; ok {
		lustre.ServerName = val
	}
	if val, ok := volParam[paramBaseDir]; ok {
		lustre.MountPoint = val
	}
	if val, ok := volParam[paramSubDir]; ok {
		lustre.SubDir = val
	}
	if val, ok := volParam[paramOnDelete]; ok {
		lustre.OnDelete = val
	}
	if val, ok := volParam[paramDIRPid]; ok {
		lustre.ProjectId = val
	}
	if val, ok := volParam[paramDIRUid]; ok {
		lustre.Uid = val
	}
}

func (cs *ControllerServer) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	volID := req.GetVolumeId()
	klog.V(4).InfoS("DeleteVolume: called", "volumeId", volID)

	if len(volID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID not provided")
	}

	// Perform any necessary cleanup here
	klog.V(4).InfoS("DeleteVolume: volume deleted successfully", "volumeId", volID)

	return &csi.DeleteVolumeResponse{}, nil
}

func (cs *ControllerServer) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "ControllerPublishVolume is not implemented")
}

func (cs *ControllerServer) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "ControllerUnpublishVolume is not implemented")
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
	return &csi.ControllerGetCapabilitiesResponse{
		Capabilities: cs.Driver.Cscap,
	}, nil
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

func getVolumeIDFromLustreVol(vol *Lustre) string {
	idElements := make([]string, totalIDElements)
	idElements[idServer] = strings.Trim(vol.ServerName, "/")
	idElements[idSubDir] = strings.Trim(vol.SubDir, "/")
	if strings.EqualFold(vol.OnDelete, retain) || strings.EqualFold(vol.OnDelete, archive) {
		idElements[idOnDelete] = vol.OnDelete
	}

	return strings.Join(idElements, separator)
}

func (cs *ControllerServer) internalMount(ctx context.Context, l *Lustre) error {
	if l.FSId == "" {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("%v is a required parameter", l.FSId))
	}
	if l.ServerName == "" {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("%v is a required parameter", l.ServerName))
	}
	if l.MountPoint == "" {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("%v is a required parameter", l.MountPoint))
	}
	// Check if the target is already a mount point
	isMountPoint, err := l.Mount.IsLikelyNotMountPoint(l.MountPoint)
	if err != nil {
		return err
	}
	if !isMountPoint {
		return nil
	}

	// Perform the mount operation
	klog.V(4).InfoS("Mounting volume", "volumeId", l.FSId, "target", l.MountPoint)
	return nil
}

func getInternalMountPath(l *Lustre) string {
	return fmt.Sprintf("%s/%s", l.MountPoint, l.SubDir)
}
