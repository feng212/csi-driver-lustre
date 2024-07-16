package main

import (
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/mount-utils"
	"os"
	"strings"
)

type Lustre struct {
	CapacityGiB int64
	MountPoint  string
	MountName   string
	StorageType string
	ProjectId   string
	Uid         string
	mount.Interface
}

func (l *Lustre) LustreMount() error {
	if l.MountName == "" {
		return status.Error(codes.InvalidArgument, "ServerName is a required parameter")
	}
	if l.MountPoint == "" {
		return status.Error(codes.InvalidArgument, "MountPoint is a required parameter")
	}

	// Check if the target is already a mount point
	isNotMountPoint, err := l.IsLikelyNotMountPoint(l.MountPoint)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(l.MountPoint, 0755); err != nil {
				return status.Error(codes.Internal, err.Error())
			}
			isNotMountPoint = true
		} else {
			return status.Error(codes.Internal, err.Error())
		}
	}

	if !isNotMountPoint {
		return nil
	}

	source := l.MountName
	target := l.MountPoint
	typefs := l.StorageType
	err = l.Mount(source, target, typefs, nil)
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

func (l *Lustre) Unmount() error {
	return nil
}

func main() {
	lustre := &Lustre{
		CapacityGiB: 1024,
		MountPoint:  "/mnt/lustre",
		MountName:   "192.168.136.11@tcp:/demo",
		StorageType: "lustre",
		Interface:   mount.New(""), // 初始化 mount.Interface
	}
	if err := lustre.LustreMount(); err != nil {
		fmt.Printf("Error mounting Lustre: %v\n", err)
	}
}
