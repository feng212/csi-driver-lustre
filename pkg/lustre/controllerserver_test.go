package lustre

import (
	"context"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"reflect"
	"testing"
)

func initTestController(_ *testing.T) *ControllerServer {
	controller := &ControllerServer{
		Driver: new(Driver),
	}
	return controller
}

func TestCreateVolume(t *testing.T) {

	testCases := []struct {
		name string
		req  *csi.CreateVolumeRequest
		resp *csi.CreateVolumeResponse
	}{
		{
			name: "test",
			req: &csi.CreateVolumeRequest{
				Name: "a1",
				VolumeCapabilities: []*csi.VolumeCapability{
					{
						AccessType: &csi.VolumeCapability_Mount{
							Mount: &csi.VolumeCapability_MountVolume{},
						},
						AccessMode: &csi.VolumeCapability_AccessMode{
							Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
						},
					},
				},
				Parameters: map[string]string{
					paramFsType:  "lustre",
					paramServer:  "172.16.100.189@tcp:/testfs",
					paramBaseDir: "/mnt/testfs",
				},
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			// Setup
			cs := initTestController(t)
			resp, err := cs.CreateVolume(context.Background(), test.req)
			// Verify
			if err != nil {
				t.Errorf("test %q failed: %v", test.name, err)
			}

			if !reflect.DeepEqual(resp, test.resp) {
				t.Errorf("test %q failed: got resp %+v, expected %+v", test.name, resp, test.resp)
			}

		})
	}
}
