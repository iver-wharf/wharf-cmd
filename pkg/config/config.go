package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/iver-wharf/wharf-core/v2/pkg/config"
	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
)

// Config holds all configurable settings for wharf-api.
//
// The config is read in the following order:
//
// 1. File: ~/.config/iver-wharf/wharf-cmd/wharf-cmd-config.yml
//
// 2. File: ./wharf-cmd-config.yml
//
// 3. File from environment variable: WHARF_CONFIG
//
// 4. Environment variables, prefixed with WHARF
//
// Each inner struct is represented as a deeper field in the different
// configurations. For YAML they represent deeper nested maps. For environment
// variables they are joined together by underscores.
//
// All environment variables must be uppercased, while YAML files are
// case-insensitive. Keeping camelCasing in YAML config files is recommended
// for consistency.
type Config struct {
	Worker         WorkerConfig
	Provisioner    ProvisionerConfig
	ProvisionerAPI ProvisionerAPIConfig
	Aggregator     AggregatorConfig
}

func (c Config) print() {
	data, err := yaml.Marshal(&c)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s\n", string(data))
}

// WorkerConfig holds settings for the worker.
type WorkerConfig struct {
	K8s   K8sConfig
	Steps StepsConfig
}

// K8sConfig holds settings for when the provisioner is using kubernetes.
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

// StepsConfig holds settings for the different types of steps.
type StepsConfig struct {
	Docker  DockerStepConfig
	Kubectl KubectlStepConfig
	Helm    HelmStepConfig
}

// DockerStepConfig holds settings for the docker step type.
type DockerStepConfig struct {
	// Image is the image for the kaniko executor to use in the docker step.
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
	// There's no config value for the Docker image tag to use, as
	// that comes from the helmVersion field in the helm step type
	// from the .wharf-ci.yml file.
	//
	// Added in v0.8.0.
	Image string
}

// ProvisionerConfig holds settings for the provisioner.
type ProvisionerConfig struct {
	K8s    K8sConfig
	Worker WorkerPodConfig
}

// WorkerPodConfig holds settings for worker pods that are created by the
// provisioner.
type WorkerPodConfig struct {
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

// ProvisionerAPIConfig holds settings for the Provisioner API.
type ProvisionerAPIConfig struct {
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

// AggregatorConfig holds settings for the aggregator.
type AggregatorConfig struct {
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

// DefaultConfig is the hard-coded default values for wharf-cmd's configs.
var DefaultConfig = Config{
	Worker: WorkerConfig{
		K8s: K8sConfig{
			Context:   "",
			Namespace: "default",
		},
		Steps: StepsConfig{
			Docker: DockerStepConfig{
				Image: "gcr.io/kaniko-project/executor:v1.7.0",
			},
			Kubectl: KubectlStepConfig{
				Image: "docker.io/wharfse/kubectl:v1.23.5",
			},
			Helm: HelmStepConfig{
				Image: "docker.io/wharfse/helm:v3.8.1",
			},
		},
	},
	Provisioner: ProvisionerConfig{
		K8s: K8sConfig{
			Context:   "",
			Namespace: "default",
		},
		Worker: WorkerPodConfig{
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
	ProvisionerAPI: ProvisionerAPIConfig{
		HTTP: HTTPConfig{
			CORS: CORSConfig{
				AllowAllOrigins: false,
				AllowOrigins:    []string{},
			},
			BindAddress: "0.0.0.0:5009",
		},
		K8s: K8sConfig{
			Context:   "",
			Namespace: "default",
		},
	},
	Aggregator: AggregatorConfig{
		WharfAPIURL:            "http://wharf-api:8080",
		WharfCMDProvisionerURL: "http://wharf-cmd-provisioner:8080",
	},
}

// LoadConfig looks for, parses and validates the config and returns it as a
// Config object.
func LoadConfig() (Config, error) {
	cfgBuilder := config.NewBuilder(DefaultConfig)

	cfgBuilder.AddConfigYAMLFile("~/.config/iver-wharf/wharf-cmd/wharf-cmd-config.yml")
	cfgBuilder.AddConfigYAMLFile(".wharf-cmd-config.yml")
	if cfgFile, ok := os.LookupEnv("WHARF_CONFIG"); ok {
		cfgBuilder.AddConfigYAMLFile(cfgFile)
	}
	cfgBuilder.AddEnvironmentVariables("WHARF")

	var cfg Config
	err := cfgBuilder.Unmarshal(&cfg)
	if err != nil {
		fmt.Printf("Failed unmarshaling: %v\n", err)
		return Config{}, err
	}
	if err := cfg.validate(); err != nil {
		fmt.Printf("Failed validating: %v\n", err)
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) validate() error {
	initContainerPullPolicy := c.Provisioner.Worker.InitContainer.ImagePullPolicy
	if ok := validateImagePullPolicy(&initContainerPullPolicy); !ok {
		return fmt.Errorf("invalid pull policy: provisioner.worker.initContainer.imagePullPolicy=%s", initContainerPullPolicy)
	}
	containerPullPolicy := c.Provisioner.Worker.Container.ImagePullPolicy
	if ok := validateImagePullPolicy(&containerPullPolicy); !ok {
		return fmt.Errorf("invalid pull policy: provisioner.worker.container.imagePullPolicy=%s", containerPullPolicy)
	}

	return nil
}

func validateImagePullPolicy(p *v1.PullPolicy) bool {
	switch strings.ToLower(string(*p)) {
	case "always":
		*p = v1.PullAlways
	case "never":
		*p = v1.PullNever
	case "ifnotpresent":
		*p = v1.PullIfNotPresent
	}
	switch *p {
	case v1.PullAlways, v1.PullIfNotPresent, v1.PullNever:
		return true
	}
	return false
}
