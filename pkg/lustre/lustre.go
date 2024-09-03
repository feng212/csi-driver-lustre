package lustre

import (
	"github.com/container-storage-interface/spec/lib/go/csi"
	"k8s.io/mount-utils"
)

const (
	DefaultDriverName = "lustre.csi.k8s.io"
	// Address of the NFS server
	// Base directory of the NFS server to create volumes under.
	// The base directory must be a direct child of the root directory.
	// The root directory is omitted from the string, for example:
	//     "base" instead of "/base"

	paramFsType          = "lustre"
	paramServer          = "server"
	paramBaseDir         = "base_dir"
	paramSubDir          = "subdir"
	paramOnDelete        = "ondelete"
	paramDIRPid          = "projectId"
	paramDIRUid          = "Uid"
	pvcNameKey           = "csi.storage.k8s.io/pvc/name"
	pvcNamespaceKey      = "csi.storage.k8s.io/pvc/namespace"
	pvNameKey            = "csi.storage.k8s.io/pv/name"
	pvcNameMetadata      = "${pvc.metadata.name}"
	pvcNamespaceMetadata = "${pvc.metadata.namespace}"
	pvNameMetadata       = "${pv.metadata.name}"
)

type DriverOptions struct {
	NodeID                       string
	DriverName                   string
	Endpoint                     string
	MountPermissions             uint64
	WorkingMountDir              string
	DefaultOnDeletePolicy        string
	VolStatsCacheExpireInMinutes int
}

type Driver struct {
	name                         string
	nodeID                       string
	version                      string
	endpoint                     string
	mountPermissions             uint64
	workingMountDir              string
	defaultOnDeletePolicy        string
	volumeLocks                  *InFlight
	is                           *IdentityServer
	ns                           *NodeServer
	cs                           *ControllerServer
	cscap                        []*csi.ControllerServiceCapability
	nscap                        []*csi.NodeServiceCapability
	vc                           []*csi.VolumeCapability_AccessMode
	VolStatsCacheExpireInMinutes int
}

type Lustre struct {
	FSId        string
	CapacityGiB int64
	SubDir      string
	MountPoint  string
	ServerName  string
	StorageType string
	ProjectId   string
	Uid         string
	Gid         string
	OnDelete    string
	Mount       mount.Interface
}
