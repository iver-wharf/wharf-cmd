package aggregator

// Config holds settings for the aggregator.
type Config struct {
	K8s K8sConfig

	// WharfAPIURL is the URL used to connect to Wharf API.
	//
	// Added in v0.8.0.
	WharfAPIURL string

	// WharfCMDProvisionerURL is the URL used to connect to the Wharf CMD
	// provisioner.
	//
	// Added in v0.8.0.
	WharfCMDProvisionerURL string
}

// K8sConfig holds settings when the worker is using kubernetes.
type K8sConfig struct {
	// Context is the context used when talking to kubernetes.
	Context string
	// Namespace is the kubernetes namespace to talk to.
	Namespace string
}
