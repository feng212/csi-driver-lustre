package main

import (
	"flag"
	"github.com/feng212/csi-driver-lustre/pkg/lustre"
	"k8s.io/klog/v2"
	"os"
)

var (
	endpoint                     = flag.String("endpoint", "unix://tmp/csi.sock", "CSI endpoint")
	nodeID                       = flag.String("nodeid", "", "node id")
	mountPermissions             = flag.Uint64("mount-permissions", 0750, "mounted folder permissions")
	driverName                   = flag.String("drivername", lustre.DefaultDriverName, "name of the driver")
	workingMountDir              = flag.String("working-mount-dir", "/tmp", "working directory for provisioner to mount lustre shares temporarily")
	defaultOnDeletePolicy        = flag.String("default-ondelete-policy", "delete", "default policy for deleting subdirectory when deleting a volume")
	volStatsCacheExpireInMinutes = flag.Int("vol-stats-cache-expire-in-minutes", 10, "The cache expire time in minutes for volume stats cache")
)

func main() {
	klog.InitFlags(nil)
	_ = flag.Set("logtostderr", "true")
	flag.Parse()
	if *nodeID == "" {
		klog.Warning("nodeid is empty")
	}

	handle()
	os.Exit(0)
}

func handle() {
	driverOptions := lustre.DriverOptions{
		NodeID:                       *nodeID,
		DriverName:                   *driverName,
		Endpoint:                     *endpoint,
		MountPermissions:             *mountPermissions,
		WorkingMountDir:              *workingMountDir,
		DefaultOnDeletePolicy:        *defaultOnDeletePolicy,
		VolStatsCacheExpireInMinutes: *volStatsCacheExpireInMinutes,
	}
	d := lustre.NewDriver(&driverOptions)
	d.Run(false)
}
