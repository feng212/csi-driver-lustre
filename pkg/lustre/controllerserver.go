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
	volumeContextSubName          = "mountname"
	idServer                      = iota
	idBaseDir
	idSubDir
	idUUID
	idOnDelete
	totalIDElements // Always last // Always last
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
	//if ok := cs.Driver.volumeLocks.Insert(volName); !ok {
	//	msg := fmt.Sprintf("Create volume request for %s is already in progress", volName)
	//	return nil, status.Error(codes.Aborted, msg)
	//}
	reqCapacity := req.GetCapacityRange().GetRequiredBytes()
	if reqCapacity == 0 {
		reqCapacity = DefaultVolumeSize
	}
	lustre := &Lustre{Mount: mount.New("")}
	lustre.OnDelete = cs.Driver.defaultOnDeletePolicy
	volParam := req.GetParameters()
	if volParam == nil {
		volParam = make(map[string]string)
	}
	lustre.StorageType = paramFsType
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
	if lustre.SubDir == "" {
		lustre.SubDir = volName
	}
	if err := validateOnDeleteValue(lustre.OnDelete); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	lustre.FSId = getVolumeIDFromLustreVol(lustre)
	err := cs.internalMount(ctx, lustre)
	if err != nil {
		return nil, status.Error(codes.Internal, "mount failed"+err.Error())
	}
	//defer func() {
	//	if err = cs.internalUnmount(ctx, nfsVol); err != nil {
	//		klog.Warningf("failed to unmount nfs server: %v", err.Error())
	//	}
	//}()
	internalVolumePath := getInternalMountPath(lustre)
	if err = os.MkdirAll(internalVolumePath, 0777); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to make subdirectory: %v", err.Error())
	}
	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      lustre.FSId,
			CapacityBytes: reqCapacity, // by setting it to zero, Provisioner will use PVC requested size as PV size
			VolumeContext: volParam,
			ContentSource: req.GetVolumeContentSource(),
		},
	}, nil
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

func getVolumeIDFromLustreVol(vol *Lustre) string {
	idElements := make([]string, totalIDElements)
	idElements[idServer] = strings.Trim(vol.ServerName, "/")
	idElements[idBaseDir] = strings.Trim(vol.MountPoint, "/")
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
		if os.IsNotExist(err) {
			if err := os.MkdirAll(l.MountPoint, 0755); err != nil {
				return status.Error(codes.Internal, err.Error())
			}
			isMountPoint = true
		} else {
			return status.Error(codes.Internal, err.Error())
		}
	}
	if !isMountPoint {
		return nil
	}
	err = l.Mount.Mount(l.ServerName, l.MountPoint, l.StorageType, nil)
	if err != nil {
		if os.IsPermission(err) {
			return status.Error(codes.PermissionDenied, err.Error())
		}
		if strings.Contains(err.Error(), "invalid argument") {
			return status.Error(codes.InvalidArgument, err.Error())
		}
		return status.Error(codes.Internal, err.Error())
	}
	return nil
}

//func (cs *ControllerServer) internalUMount(ctx context.Context, l *Lustre) error {
//	extensiveMountPointCheck := true
//	forceUnmounter, ok := l.Mount.(mount.MounterForceUnmounter)
//	if ok {
//		klog.V(2).Infof("force unmount %s on %s", volumeID, targetPath)
//		err := mount.CleanupMountWithForce(targetPath, forceUnmounter, extensiveMountPointCheck, 30*time.Second)
//	} else {
//		err = mount.CleanupMountPoint(targetPath, ns.mounter, extensiveMountPointCheck)
//	}
//	if err != nil {
//		return nil, status.Errorf(codes.Internal, "failed to unmount target %q: %v", targetPath, err)
//	}
//	klog.V(2).Infof("NodeUnpublishVolume: unmount volume %s on %s successfully", volumeID, targetPath)
//
//}

func getInternalMountPath(vol *Lustre) string {
	if vol == nil {
		return ""
	}
	return vol.MountPoint + "/" + vol.SubDir
}
