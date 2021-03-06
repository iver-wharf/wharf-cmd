package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/iver-wharf/wharf-core/v2/pkg/config"
	v1 "k8s.io/api/core/v1"
)

// Config holds all configurable settings for wharf-api.
//
// The config is read in the following order:
//
// 1. File: /etc/iver-wharf/wharf-cmd/wharf-cmd-config.yml
//
// 2. File: (config home)/iver-wharf/wharf-cmd/wharf-cmd-config.yml, depending
// on OS:
//
//  - Linux:       ~/.config/iver-wharf/wharf-cmd/wharf-cmd-config.yml
//  - Darwin/OS X: ~/Library/Application Support/iver-wharf/wharf-cmd/wharf-cmd-config.yml
//  - Windows:     %APPDATA%\iver-wharf\wharf-cmd\wharf-cmd-config.yml
//
// 2. File: ./wharf-cmd-config.yml
//
// 3. File from environment variable: WHARF_CONFIG
//
// 4. Environment variables, prefixed with WHARF_
//
// Each inner struct is represented as a deeper field in the different
// configurations. For YAML they represent deeper nested maps. For environment
// variables they are joined together by underscores.
//
// All environment variables must be uppercased, while YAML files are
// case-insensitive. Keeping camelCasing in YAML config files is recommended
// for consistency.
type Config struct {
	// K8s holds Kubernetes-specific settings.
	//
	// Added in v0.8.0.
	K8s K8sConfig
	// Worker holds settings specific to the wharf-cmd-worker.
	//
	// Added in v0.8.0.
	Worker WorkerConfig
	// Provisioner holds settings specific to the wharf-cmd-provisioner.
	//
	// Added in v0.8.0.
	Provisioner ProvisionerConfig
	// Aggregator holds settings specific to the wharf-cmd-aggregator.
	//
	// Added in v0.8.0.
	Aggregator AggregatorConfig
	// Watchdog holds settings specific to the wharf-cmd-watchdog.
	//
	// Added in v0.8.0.
	Watchdog WatchdogConfig

	// InstanceID may be an arbitrary string that is used to identify different
	// Wharf installations from each other. Needed when you use multiple Wharf
	// installations in the same environment, such as the same Kubernetes
	// namespace, to let Wharf know which builds belong to which Wharf
	// installation.
	//
	// Added in v0.8.0.
	InstanceID string
}

// K8sConfig holds settings for using kubernetes.
type K8sConfig struct {
	// Context is the context used when talking to kubernetes.
	//
	// Added in v0.8.0.
	Context string
	// Namespace is the kubernetes namespace to talk to.
	//
	// Added in v0.8.0.
	Namespace string
}

// WorkerConfig holds settings for the worker.
type WorkerConfig struct {
	// Steps holds settings specific to all different step types.
	//
	// Added in v0.8.0.
	Steps StepsConfig
}

// StepsConfig holds settings for the different types of steps.
type StepsConfig struct {
	// Docker holds settings for the docker step type. (Building docker images)
	//
	// Added in v0.8.0.
	Docker DockerStepConfig
	// Kubectl holds settings for the kubectl step type. (Applying Kubernetes
	// YAML manifests)
	//
	// Added in v0.8.0.
	Kubectl KubectlStepConfig
	// Helm holds settings for the helm step type. (Installing or upgrading
	// Helm releases)
	//
	// Added in v0.8.0.
	Helm HelmStepConfig

	// TODO: Add ContainerStepConfig, see pkg/worker/k8spodtemplating.go
	// TODO: Add HelmPackageStepConfig, see pkg/worker/k8spodtemplating.go
	// TODO: Add NuGetPackageStepConfig, see pkg/worker/k8spodtemplating.go
}

// DockerStepConfig holds settings for the docker step type.
type DockerStepConfig struct {
	// Image is the image for the kaniko executor to use.
	//
	// Added in v0.8.0.
	Image string
	// ImageTag is the image tag to use for the image.
	//
	// Added in v0.8.0.
	ImageTag string
}

// KubectlStepConfig holds settings for the kubectl step type.
type KubectlStepConfig struct {
	// Image is the image to use in the kubectl step.
	//
	// Added in v0.8.0.
	Image string
	// ImageTag is the image tag to use for the image.
	//
	// Added in v0.8.0.
	ImageTag string
}

// HelmStepConfig holds settings for the helm step type.
type HelmStepConfig struct {
	// Image is the image to use in the helm step.
	//
	// There's no config value for the Docker image tag to use, as that comes
	// from the helmVersion field in the helm step type from the .wharf-ci.yml
	// file.
	//
	// Added in v0.8.0.
	Image string
}

// ProvisionerConfig holds settings for the provisioner.
type ProvisionerConfig struct {
	// HTTP holds settings for the wharf-cmd-provisioner's HTTP server
	// (gRPC & REST).
	//
	// Added in v0.8.0.
	HTTP HTTPConfig
	// K8s holds settings for how the wharf-cmd-provisioner acts towards
	// Kubernetes.
	//
	// Added in v0.8.0.
	K8s ProvisionerK8sConfig
}

// ProvisionerK8sConfig holds kubernetes specific settings for the provisioner.
type ProvisionerK8sConfig struct {
	// Worker holds settings for how the wharf-cmd-provisioner provisions
	// wharf-cmd-workers to Kubernetes.
	//
	// Added in v0.8.0.
	Worker ProvisionerK8sWorkerConfig
}

// ProvisionerK8sWorkerConfig holds settings for worker pods that are created by the
// provisioner.
type ProvisionerK8sWorkerConfig struct {
	// ServiceAccountName is the service account name to use for the pod.
	//
	// Added in v0.8.0.
	ServiceAccountName string
	// InitContainer holds settings for the init container.
	//
	// Added in v0.8.0.
	InitContainer K8sContainerConfig
	// Container holds settings for the container.
	//
	// Added in v0.8.0.
	Container K8sContainerConfig
	// ImagePullSecrets is an optional list of references to secrets in the same
	// namespace to use for pulling either the wharf-cmd-worker init container
	// or app container.
	//
	// More info: https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod
	//
	// Added in v0.8.0.
	ImagePullSecrets []K8sLocalObjectReference
	// ConfigMapName is the name of a Kubernetes ConfigMap to mount into the
	// wharf-cmd-worker Pod. The ConfigMap can have the following keys, that
	// corresponds to their respective files:
	//
	//  wharf-cmd-config.yml
	//  wharf-vars.yml
	//
	// Sample ConfigMap manifest:
	//
	//  apiVersion: v1
	//  kind: ConfigMap
	//  metadata:
	//    name: wharf-cmd-worker-config
	//  data:
	//    wharf-cmd-config.yml: |
	//      instanceId: dev
	//    wharf-vars.yml: |
	//      vars:
	//        REG_SECRET: docker-registry
	//
	// With the above manifest created in the same namespace as where the
	// wharf-cmd-worker pods are created, you would then set this field to
	//
	//  wharf-cmd-worker-config
	//
	// Added in v0.8.0.
	ConfigMapName string
	// ExtraEnvs is an array of additional environment variables to set on the
	// wharf-cmd-worker Pod.
	//
	// Added in v0.8.0.
	ExtraEnvs []K8sEnvVar
	// ExtraArgs is an array of additional command-line arguments to add to the
	// wharf-cmd-worker Pod.
	//
	// Added in v0.8.0.
	ExtraArgs []string
}

// K8sContainerConfig holds settings for a kubernetes container.
type K8sContainerConfig struct {
	// Image is the base path for the image to use for a container.
	//
	// Added in v0.8.0.
	Image string
	// ImageTag is the version tag to use for a container.
	//
	// Added in v0.8.0.
	ImageTag string
	// ImagePullPolicy is the image pull policy to use for a container.
	//
	// "Always"
	//   Attempts to pull the latest image.
	//   Container will fail if the pull fails.
	//
	// "Never"
	//   Never pulls an image, i.e. only uses a local image.
	//   Container will fail if the image isn't present
	//
	// "IfNotPresent"
	//   Pulls if the image isn't present on disk.
	//   Container will fail if the image isn't present and the pull fails.
	//
	// Added in v0.8.0.
	ImagePullPolicy v1.PullPolicy
}

// HTTPConfig holds settings for the HTTP server.
type HTTPConfig struct {
	// CORS holds settings regarding HTTP CORS (Cross-Origin Resource Sharing).
	//
	// Added in v0.8.0.
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

// AggregatorConfig holds settings for the aggregator.
type AggregatorConfig struct {
	// WharfAPIURL is the URL used to connect to Wharf API.
	//
	// Added in v0.8.0.
	WharfAPIURL string
	// WorkerAPIExternalPort is the port used to connect to a Wharf worker.
	//
	// Added in v0.8.0.
	WorkerAPIExternalPort int16
}

// WatchdogConfig holds settings for the watchdog.
type WatchdogConfig struct {
	// WharfAPIURL is the URL used to connect to Wharf API.
	//
	// Added in v0.8.0.
	WharfAPIURL string
	// ProvisionerURL is the URL used to connect to the Wharf Cmd
	// provisioner.
	//
	// Added in v0.8.0.
	ProvisionerURL string
}

// DefaultConfig is the hard-coded default values for wharf-cmd's configs.
var DefaultConfig = Config{
	InstanceID: "local",
	K8s: K8sConfig{
		Context:   "",
		Namespace: "",
	},
	Worker: WorkerConfig{
		Steps: StepsConfig{
			Docker: DockerStepConfig{
				Image:    "gcr.io/kaniko-project/executor",
				ImageTag: "v1.7.0",
			},
			Kubectl: KubectlStepConfig{
				Image:    "docker.io/wharfse/kubectl",
				ImageTag: "v1.23.5",
			},
			Helm: HelmStepConfig{
				Image: "docker.io/wharfse/helm",
			},
		},
	},
	Provisioner: ProvisionerConfig{
		HTTP: HTTPConfig{
			CORS: CORSConfig{
				AllowAllOrigins: false,
				AllowOrigins:    []string{},
			},
			BindAddress: "0.0.0.0:5009",
		},
		K8s: ProvisionerK8sConfig{
			Worker: ProvisionerK8sWorkerConfig{
				ServiceAccountName: "wharf-cmd",
				InitContainer: K8sContainerConfig{
					Image:           "bitnami/git",
					ImageTag:        "2-debian-10",
					ImagePullPolicy: v1.PullIfNotPresent,
				},
				Container: K8sContainerConfig{
					Image:           "quay.io/iver-wharf/wharf-cmd",
					ImageTag:        "latest",
					ImagePullPolicy: v1.PullAlways,
				},
			},
		},
	},
	Aggregator: AggregatorConfig{
		WharfAPIURL:           "http://localhost:5001",
		WorkerAPIExternalPort: 5010,
	},
	Watchdog: WatchdogConfig{
		WharfAPIURL:    "http://localhost:5001",
		ProvisionerURL: "http://localhost:5009",
	},
}

// LoadConfig looks for, parses and validates the config and returns it as a
// Config object.
func LoadConfig() (Config, error) {
	cfgBuilder := config.NewBuilder(DefaultConfig)

	cfgBuilder.AddConfigYAMLFile("/etc/iver-wharf/wharf-cmd/wharf-cmd-config.yml")
	if confDir, err := os.UserConfigDir(); err == nil {
		cfgBuilder.AddConfigYAMLFile(filepath.Join(confDir, "iver-wharf/wharf-cmd/wharf-cmd-config.yml"))
	}
	cfgBuilder.AddConfigYAMLFile(".wharf-cmd-config.yml")
	if cfgFile, ok := os.LookupEnv("WHARF_CONFIG"); ok {
		cfgBuilder.AddConfigYAMLFile(cfgFile)
	}
	cfgBuilder.AddEnvironmentVariables("WHARF")

	var cfg Config
	if err := cfgBuilder.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("load config: %w", err)
	}
	if err := cfg.validate(); err != nil {
		return Config{}, fmt.Errorf("load config: %w", err)
	}

	return cfg, nil
}

func (c *Config) validate() error {
	w := &c.Provisioner.K8s.Worker
	var ok bool

	w.InitContainer.ImagePullPolicy, ok = parseImagePolicy(w.InitContainer.ImagePullPolicy)
	if !ok {
		return fmt.Errorf("invalid pull policy: provisioner.worker.initContainer.imagePullPolicy=%s", w.InitContainer.ImagePullPolicy)
	}

	w.Container.ImagePullPolicy, ok = parseImagePolicy(w.Container.ImagePullPolicy)
	if !ok {
		return fmt.Errorf("invalid pull policy: provisioner.worker.container.imagePullPolicy=%s", w.Container.ImagePullPolicy)
	}
	return nil
}

func parseImagePolicy(p v1.PullPolicy) (v1.PullPolicy, bool) {
	switch strings.ToLower(string(p)) {
	case "always":
		return v1.PullAlways, true
	case "never":
		return v1.PullNever, true
	case "ifnotpresent":
		return v1.PullIfNotPresent, true
	default:
		return v1.PullPolicy(""), false
	}
}
