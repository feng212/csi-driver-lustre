package lustre

import (
	"context"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
)

func initTestNode(_ *testing.T) *NodeServer {
	nodeserver := &NodeServer{
		Driver: new(Driver),
	}
	return nodeserver
}

func TestNodeServer_NodePublishVolume(t *testing.T) {
	params := map[string]string{
		paramFsType:  "lustre",
		paramServer:  "192.168.136.11@tcp:/lustre",
		paramBaseDir: "/mnt/testfs",
		paramSubDir:  "a1",
	}

	volumeCap := csi.VolumeCapability_AccessMode{Mode: csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER}

	tests := []struct {
		desc          string
		setup         func()
		req           csi.NodePublishVolumeRequest
		skipOnWindows bool
		expectedErr   error
		cleanup       func()
	}{
		{
			desc: "[Error] invalid mountPermissions",
			req: csi.NodePublishVolumeRequest{
				VolumeContext:    params,
				VolumeCapability: &csi.VolumeCapability{AccessMode: &volumeCap},
				VolumeId:         "vol_1",
				TargetPath:       "/var/lib/test",
				Readonly:         true},
			expectedErr: status.Error(codes.InvalidArgument, "invalid mountPermissions 07ab"),
		},
	}

	for _, tc := range tests {
		ns := initTestNode(t)
		ns.NodePublishVolume(context.Background(), &tc.req)

	}
}
