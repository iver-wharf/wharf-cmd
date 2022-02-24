package config

import (
	"os"

	"github.com/iver-wharf/wharf-core/pkg/config"
)

// Config holds all configurable settings for wharf-cmd.
//
// The config is read in the following order:
//
// 1. File: /etc/iver-wharf/wharf-cmd/config.yml
//
// 2. File: ./wharf-api-config.yml
//
// 3. File from environment variable: WHARF_CMD_CONFIG
//
// Each inner struct is represented as a deeper field in the different
// configurations. For YAML they represent deeper nested maps. For environment
// variables they are joined together by underscores.
type Config struct {
	WorkerHTTPServer WorkerHTTPServerConfig

	// InstanceID may be an arbitrary string that is used to identify different
	// Wharf installations from each other. Needed when you use multiple Wharf
	// installations in the same environment, such as the same Kubernetes
	// namespace or the same Jenkins instance, to let Wharf know which builds
	// belong to which Wharf installation.
	//
	// Added in v0.8.0.
	InstanceID string
}

// WorkerHTTPServerConfig holds settings for the worker HTTP server.
type WorkerHTTPServerConfig struct {
	HTTP HTTPConfig
	CA   CertConfig
}

// HTTPConfig holds settings for an HTTP server.
type HTTPConfig struct {
	CORS CORSConfig

	// BindAddress is the IP-address and port, separated by a colon, to bind
	// the HTTP server to. An IP-address of 0.0.0.0 will bind to all
	// IP-addresses.
	//
	// Added in v0.8.0.
	BindAddress string

	// BasicAuth is a comma-separated list of username:password pairs.
	//
	// Example for user named "admin" with password "1234" and a user named
	// "john" with the password "secretpass":
	// 	BasicAuth="admin:1234,john:secretpass"
	//
	// Added in v0.8.0.
	BasicAuth string
}

// CORSConfig holds settings for an HTTP server's CORS settings.
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

// CertConfig holds settings for certificates verification used when talking
// to remote services over HTTPS.
type CertConfig struct {
	// CertsFile points to a file of one or more PEM-formatted certificates to
	// use in addition to the certificates from the system
	// (such as from /etc/ssl/certs/).
	//
	// Added in v0.8.0.
	CertsFile   string
	CertKeyFile string
}

// DefaultConfig is the hard-coded default values for wharf-cmd's configs.
var DefaultConfig = Config{
	WorkerHTTPServer: WorkerHTTPServerConfig{
		HTTP: HTTPConfig{
			BindAddress: "0.0.0.0:8080",
			CORS: CORSConfig{
				AllowAllOrigins: true,
			},
		},
	},
}

// LoadConfig loads the config for wharf-cmd.
func LoadConfig() (Config, error) {
	cfgBuilder := config.NewBuilder(DefaultConfig)

	cfgBuilder.AddConfigYAMLFile("/etc/iver-wharf/wharf-cmd/config.yml")
	cfgBuilder.AddConfigYAMLFile("wharf-cmd-config.yml")
	if cfgFile, ok := os.LookupEnv("WHARF_CMD_CONFIG"); ok {
		cfgBuilder.AddConfigYAMLFile(cfgFile)
	}

	var cfg Config
	err := cfgBuilder.Unmarshal(&cfg)
	if err != nil {
		return Config{}, err
	}
	if err := cfg.validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (cfg *Config) validate() error {
	return nil
}
