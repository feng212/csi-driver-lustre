package lustre_driver

import (
	"context"
	"github.com/container-storage-interface/spec/lib/go/csi"
)

type identityService struct {
}

func (is *identityService) GetPluginInfo(ctx context.Context, req *csi.GetPluginInfoRequest) (*csi.GetPluginInfoResponse, error) {
	return nil, nil
}

func (is *identityService) GetPluginCapabilities(ctx context.Context, req *csi.GetPluginCapabilitiesRequest) (*csi.GetPluginCapabilitiesResponse, error) {
	return nil, nil
}

func (is *identityService) Probe(ctx context.Context, req *csi.ProbeRequest) (*csi.ProbeResponse, error) {
	return nil, nil
}
