package lustre_driver

import (
	"context"
	"csi-driver-lustre/pkg/lustre-driver/lustrefs"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"reflect"
	"testing"
)

func initTestController(_ *testing.T) *controllerService {
	inflight := &lustrefs.InFlight{}
	lustre := &lustrefs.Lustre{}
	controller := &controllerService{
		inflight,
		lustre,
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
				Name: "test",
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
					paramServer:      "192.168.136.11@tcp:/demo",
					paramSubDir:      "/mnt",
					paramStorageType: "lustre",
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
			if err == nil {
				t.Errorf("test %q failed; got success", test.name)
			}
			if !reflect.DeepEqual(resp, test.resp) {
				t.Errorf("test %q failed: got resp %+v, expected %+v", test.name, resp, test.resp)
			}

		})
	}
}
