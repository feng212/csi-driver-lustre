package lustre_driver

type Driver struct {
}

const (
	DefaultDriverName = "lustre.csi.k8s.io"
	// Address of the NFS server
	paramServer = "server"
	// Base directory of the NFS server to create volumes under.
	// The base directory must be a direct child of the root directory.
	// The root directory is omitted from the string, for example:
	//     "base" instead of "/base"
	paramSubDir          = "subdir"
	paramStorageType     = "lustre"
	paramDIRPid          = "projectId"
	paramDIRUid          = "Uid"
	pvcNameKey           = "csi.storage.k8s.io/pvc/name"
	pvcNamespaceKey      = "csi.storage.k8s.io/pvc/namespace"
	pvNameKey            = "csi.storage.k8s.io/pv/name"
	pvcNameMetadata      = "${pvc.metadata.name}"
	pvcNamespaceMetadata = "${pvc.metadata.namespace}"
	pvNameMetadata       = "${pv.metadata.name}"
)
