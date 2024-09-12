package lustre

import (
	"context"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"k8s.io/klog/v2"
)

type IdentityServer struct {
	csi.UnimplementedIdentityServer
	Driver *Driver
}

// GetPluginInfo 返回 CSI 驱动的基本信息
func (ids *IdentityServer) GetPluginInfo(ctx context.Context, req *csi.GetPluginInfoRequest) (*csi.GetPluginInfoResponse, error) {
	klog.V(4).InfoS("GetPluginInfo called")

	if ids.Driver.Name == "" {
		return nil, status.Error(codes.Unavailable, "Driver Name not configured")
	}

	return &csi.GetPluginInfoResponse{
		Name:          ids.Driver.Name,
		VendorVersion: ids.Driver.Version,
	}, nil
}

// GetPluginCapabilities 返回 CSI 驱动的功能，比如 Service 和 ControllerService
func (ids *IdentityServer) GetPluginCapabilities(ctx context.Context, req *csi.GetPluginCapabilitiesRequest) (*csi.GetPluginCapabilitiesResponse, error) {
	klog.V(4).InfoS("GetPluginCapabilities called")

	capabilities := []*csi.PluginCapability{
		{
			Type: &csi.PluginCapability_Service_{
				Service: &csi.PluginCapability_Service{
					Type: csi.PluginCapability_Service_CONTROLLER_SERVICE,
				},
			},
		},
		{
			Type: &csi.PluginCapability_Service_{
				Service: &csi.PluginCapability_Service{
					Type: csi.PluginCapability_Service_VOLUME_ACCESSIBILITY_CONSTRAINTS,
				},
			},
		},
	}

	return &csi.GetPluginCapabilitiesResponse{
		Capabilities: capabilities,
	}, nil
}

// Probe 用于检测 CSI 插件的健康状态
func (ids *IdentityServer) Probe(ctx context.Context, req *csi.ProbeRequest) (*csi.ProbeResponse, error) {
	klog.V(4).InfoS("Probe called")

	// 如果插件没有正确配置，可以返回非健康状态
	if ids.Driver == nil {
		return nil, status.Error(codes.FailedPrecondition, "Driver not configured")
	}

	// 插件健康，返回成功
	return &csi.ProbeResponse{Ready: &wrapperspb.BoolValue{Value: true}}, nil
}
