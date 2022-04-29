package main

import (
	"os"

	"github.com/iver-wharf/wharf-cmd/pkg/aggregator"
	"github.com/iver-wharf/wharf-cmd/pkg/provisioner"
	"github.com/iver-wharf/wharf-cmd/pkg/provisionerapi"
	"github.com/iver-wharf/wharf-cmd/pkg/worker"
	"github.com/iver-wharf/wharf-core/v2/pkg/config"
)

// Config holds all configurable settings for wharf-api.
//
// The config is read in the following order:
//
// 1. File: ~/.config/iver-wharf/wharf-cmd/wharf-cmd-config.yml
//
// 2. File: ./wharf-cmd-config.yml
//
// 3. File from environment variable: WHARF_CMD_CONFIG
//
// 4. Environment variables, prefixed with WHARF_CMD
//
// Each inner struct is represented as a deeper field in the different
// configurations. For YAML they represent deeper nested maps. For environment
// variables they are joined together by underscores.
//
// All environment variables must be uppercased, while YAML files are
// case-insensitive. Keeping camelCasing in YAML config files is recommended
// for consistency.
type Config struct {
	Worker         worker.Config
	Provisioner    provisioner.Config
	ProvisionerAPI provisionerapi.Config
	Aggregator     aggregator.Config
}

// DefaultConfig is the hard-coded default values for wharf-cmd's configs.
var DefaultConfig = Config{
	Worker: worker.Config{
		K8s: worker.K8sConfig{
			Context:   "",
			Namespace: "default",
		},
		Steps: worker.StepsConfig{
			Docker: worker.DockerStepConfig{
				KanikoImage: "gcr.io/kaniko-project/executor:v1.7.0",
			},
			Kubectl: worker.KubectlStepConfig{
				KubectlImage: "docker.io/wharfse/kubectl:v1.23.5",
			},
			Helm: worker.HelmStepConfig{
				HelmImage: "docker.io/wharfse/helm:v3.8.1",
			},
		},
	},
	Provisioner: provisioner.Config{
		K8s: provisioner.K8sConfig{
			Context:   "",
			Namespace: "default",
		},
	},
	ProvisionerAPI: provisionerapi.Config{
		HTTP: provisionerapi.HTTPConfig{
			CORS: provisionerapi.CORSConfig{
				AllowAllOrigins: false,
				AllowOrigins:    []string{},
			},
			BindAddress: "0.0.0.0:5009",
		},
		K8s: provisionerapi.K8sConfig{
			Context:   "",
			Namespace: "default",
		},
	},
	Aggregator: aggregator.Config{
		WharfAPIURL:            "http://wharf-api:8080",
		WharfCMDProvisionerURL: "http://wharf-cmd-provisioner:8080",
	},
}

func loadConfig() (Config, error) {
	cfgBuilder := config.NewBuilder(DefaultConfig)

	cfgBuilder.AddConfigYAMLFile("~/.config/iver-wharf/wharf-cmd/wharf-cmd-config.yml")
	cfgBuilder.AddConfigYAMLFile("wharf-cmd-config.yml")
	if cfgFile, ok := os.LookupEnv("WHARF_CMD_CONFIG"); ok {
		cfgBuilder.AddConfigYAMLFile(cfgFile)
	}
	cfgBuilder.AddEnvironmentVariables("WHARF_CMD")

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
