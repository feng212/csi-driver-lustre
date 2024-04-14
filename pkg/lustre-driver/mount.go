package lustre_driver

import (
	"k8s.io/mount-utils"
	"os"
)

type Mounter interface {
	mount.Interface
	IsCorruptedMnt(err error) bool
	PathExists(path string) (bool, error)
	MakeDir(pathname string) error
}

type NodeMounter struct {
	mount.Interface
}

func newNodeMounter() (Mounter, error) {
	return &NodeMounter{
		Interface: mount.New(""),
	}, nil
}

func (n NodeMounter) IsCorruptedMnt(err error) bool {
	return mount.IsCorruptedMnt(err)
}

func (n NodeMounter) PathExists(path string) (bool, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func (n NodeMounter) MakeDir(pathname string) error {
	err := os.MkdirAll(pathname, os.FileMode(0755))
	if err != nil {
		if !os.IsExist(err) {
			return err
		}
	}
	return nil
}
