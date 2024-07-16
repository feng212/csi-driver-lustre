package lustre_driver

import (
	"context"
	"csi-driver-lustre/pkg/lustre-driver/lustrefs"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"k8s.io/mount-utils"
	"reflect"
	"testing"
)

func initTestController(_ *testing.T) *controllerService {

	lustre := &lustrefs.Lustre{Interface: mount.New("")}
	controller := &controllerService{
		lustrefs.NewInFlight(),
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
					paramServer:      "10.10.8.131@o2ib:/testfs",
					paramBaseDir:     "/mnt/testfs",
					paramSubDir:      "34fsdf-df6er2-3",
					paramStorageType: "wistor",
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
