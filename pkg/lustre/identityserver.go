package lustre

import "github.com/container-storage-interface/spec/lib/go/csi"

type IdentityServer struct {
	csi.UnimplementedControllerServer
}
