package lustre

import (
	"github.com/container-storage-interface/spec/lib/go/csi"
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
