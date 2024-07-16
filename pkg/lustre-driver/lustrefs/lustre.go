package lustrefs

import (
	"bytes"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/mount-utils"
	"math/rand/v2"
	"os"
	"os/exec"
	"strings"
)

const (
	DefaultVolumeSize int64 = 1200
	idServer                = iota
	idBaseDir
	idSubDir
	idUUID
	idOnDelete
	totalIDElements // Always last
)

type FS interface {
	CreateFs()
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
	mount.Interface
}

func (l Lustre) CreateFs() error {
	err := l.MountFS()
	if err != nil {
		return err
	}
	err = l.CreateDirFs()
	if err != nil {
		return err
	}
	l.FSId = l.getVolumeIDFromNfsVol()
	return nil
}

func (l *Lustre) MountFS() error {
	if l.ServerName == "" {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("%v is a required parameter", l.ServerName))
	}
	if l.MountPoint == "" {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("%v is a required parameter", l.MountPoint))
	}
	// Check if the target is already a mount point
	isMountPoint, err := l.IsLikelyNotMountPoint(l.MountPoint)
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
	err = l.Mount(l.ServerName, l.MountPoint, l.StorageType, nil)
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

func (l *Lustre) CreateDirFs() error {
	_, err := os.Stat(l.MountPoint + "/" + l.SubDir)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.Mkdir(l.MountPoint+"/"+l.SubDir, 0755); err != nil {
				return status.Error(codes.Internal, err.Error())
			}
		}
	}
	// set user
	if l.Uid != "" {
		cmd := exec.Command("setfacl", "-m", "u:"+l.Uid+":rwx", l.MountPoint+"/"+l.SubDir)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
		fmt.Printf("out:\n%s\n err:\n%s\n", outStr, errStr)
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}
	}

	if l.Gid != "" {
		cmd := exec.Command("setfacl", "-m", "g:"+l.Gid+":rwx", l.MountPoint+"/"+l.SubDir)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
		fmt.Printf("out:\n%s\n err:\n%s\n", outStr, errStr)
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}
	}
	return nil
}

func (l *Lustre) Unmount() error {
	return nil
}

func (l *Lustre) getVolumeIDFromNfsVol() string {
	idElements := make([]string, totalIDElements)
	idElements[idServer] = strings.Trim(l.ServerName, "/")
	idElements[idBaseDir] = strings.Trim(l.MountPoint, "/")
	idElements[idSubDir] = strings.Trim(l.SubDir, "/")
	return strings.Join(idElements, fmt.Sprintf("fs-%d", rand.Uint64()))
}
