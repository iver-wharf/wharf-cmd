package provisionerapi

// Config holds settings for the Provisioner API.
type Config struct {
	HTTP HTTPConfig
	K8s  K8sConfig
}

// HTTPConfig holds settings for the HTTP server.
type HTTPConfig struct {
	CORS CORSConfig

	// BindAddress is the IP-address and port, separated by a colon, to bind
	// the HTTP server to. An IP-address of 0.0.0.0 will bind to all
	// IP-addresses.
	//
	// Added in v0.8.0.
	BindAddress string
}

// CORSConfig holds settings for the HTTP server's CORS settings.
type CORSConfig struct {
	// AllowAllOrigins enables CORS and allows all hostnames and URLs in the
	// HTTP request origins when set to true. Practically speaking, this
	// results in the HTTP header "Access-Control-Allow-Origin" set to "*".
	//
	// Added in v0.8.0.
	AllowAllOrigins bool

	// AllowOrigins enables CORS and allows the list of origins in the
	// HTTP request origins when set. Practically speaking, this
	// results in the HTTP header "Access-Control-Allow-Origin".
	//
	// Added in v0.8.0.
	AllowOrigins []string
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
