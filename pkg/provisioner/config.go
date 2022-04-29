package provisioner

// Config holds settings for the provisioner.
type Config struct {
	K8s K8sConfig
}

// K8sConfig holds settings for when the provisioner is using kubernetes.
type K8sConfig struct {
	// Context is the context used when talking to kubernetes.
	//
	// Added in v0.8.0
	Context string

	// Namespace is the kubernetes namespace to talk to.
	//
	// Added in v0.8.0
	Namespace string
}
