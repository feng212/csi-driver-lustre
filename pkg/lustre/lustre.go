package lustre

import (
	"github.com/container-storage-interface/spec/lib/go/csi"
	"k8s.io/klog/v2"
	"k8s.io/mount-utils"
	"runtime"
	azcache "sigs.k8s.io/cloud-provider-azure/pkg/cache"
	"time"
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
	pvcNameKey           = "csi.storage.k8s.io/pvc.yaml/Name"
	pvcNamespaceKey      = "csi.storage.k8s.io/pvc.yaml/namespace"
	pvNameKey            = "csi.storage.k8s.io/pv/Name"
	pvcNameMetadata      = "${pvc.yaml.metadata.Name}"
	pvcNamespaceMetadata = "${pvc.yaml.metadata.namespace}"
	pvNameMetadata       = "${pv.metadata.Name}"
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
	Name                         string
	NodeId                       string
	Version                      string
	Endpoint                     string
	MountPermissions             uint64
	WorkingMountDir              string
	DefaultOnDeletePolicy        string
	VolumeLocks                  *InFlight
	Is                           *IdentityServer
	Ns                           *NodeServer
	Cs                           *ControllerServer
	Cscap                        []*csi.ControllerServiceCapability
	Nscap                        []*csi.NodeServiceCapability
	Vc                           []*csi.VolumeCapability_AccessMode
	VolStatsCache                azcache.Resource
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

func NewDriver(options *DriverOptions) *Driver {
	klog.V(2).Infof("Driver: %v version: %v", options.DriverName, driverVersion)

	n := &Driver{
		Name:                         options.DriverName,
		NodeId:                       options.NodeID,
		Version:                      driverVersion,
		Endpoint:                     options.Endpoint,
		MountPermissions:             options.MountPermissions,
		WorkingMountDir:              options.WorkingMountDir,
		DefaultOnDeletePolicy:        options.DefaultOnDeletePolicy,
		VolStatsCacheExpireInMinutes: options.VolStatsCacheExpireInMinutes,
	}
	n.AddControllerServiceCapabilities([]csi.ControllerServiceCapability_RPC_Type{
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
		csi.ControllerServiceCapability_RPC_SINGLE_NODE_MULTI_WRITER,
		csi.ControllerServiceCapability_RPC_CLONE_VOLUME,
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT,
	})

	n.AddNodeServiceCapabilities([]csi.NodeServiceCapability_RPC_Type{
		csi.NodeServiceCapability_RPC_GET_VOLUME_STATS,
		csi.NodeServiceCapability_RPC_SINGLE_NODE_MULTI_WRITER,
		csi.NodeServiceCapability_RPC_UNKNOWN,
	})
	n.VolumeLocks = NewInFlight()

	if options.VolStatsCacheExpireInMinutes <= 0 {
		options.VolStatsCacheExpireInMinutes = 10 // default expire in 10 minutes
	}

	var err error
	getter := func(key string) (interface{}, error) { return nil, nil }
	if n.VolStatsCache, err = azcache.NewTimedCache(time.Duration(options.VolStatsCacheExpireInMinutes)*time.Minute, getter, false); err != nil {
		klog.Fatalf("%v", err)
	}
	return n
}

func NewControllerServer(n *Driver) *ControllerServer {
	return &ControllerServer{
		Driver: n,
	}
}

func NewNodeServer(n *Driver, mounter mount.Interface) *NodeServer {
	return &NodeServer{
		Driver: n,
		Mount:  mounter,
	}
}

func NewDefaultIdentityServer(d *Driver) *IdentityServer {
	return &IdentityServer{
		Driver: d,
	}
}

func (n *Driver) Run(testMode bool) {
	versionMeta, err := GetVersionYAML(n.Name)
	if err != nil {
		klog.Fatalf("%v", err)
	}
	klog.V(2).Infof("\nDRIVER INFORMATION:\n-------------------\n%s\n\nStreaming logs below:", versionMeta)

	mounter := mount.New("")
	if runtime.GOOS == "linux" {
		// MounterForceUnmounter is only implemented on Linux now
		mounter = mounter.(mount.MounterForceUnmounter)
	}
	s := NewNonBlockingGRPCServer()
	s.Start(n.Endpoint,
		NewDefaultIdentityServer(n),
		// NFS plugin has not implemented ControllerServer
		// using default controllerserver.
		NewControllerServer(n),
		NewNodeServer(n, mounter),
		testMode)
	s.Wait()
}

func (n *Driver) AddControllerServiceCapabilities(cl []csi.ControllerServiceCapability_RPC_Type) {
	var csc []*csi.ControllerServiceCapability
	for _, c := range cl {
		csc = append(csc, NewControllerServiceCapability(c))
	}
	n.Cscap = csc
}

func (n *Driver) AddNodeServiceCapabilities(nl []csi.NodeServiceCapability_RPC_Type) {
	var nsc []*csi.NodeServiceCapability
	for _, n := range nl {
		nsc = append(nsc, NewNodeServiceCapability(n))
	}
	n.Nscap = nsc
}
