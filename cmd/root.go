package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/iver-wharf/wharf-core/pkg/logger"
	"github.com/iver-wharf/wharf-core/pkg/logger/consolepretty"
	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var loglevel string
var Kubeconfig *rest.Config
var kubeconfigPath string
var Namespace string

var rootCmd = &cobra.Command{
	Use:   "wharf-ci",
	Short: "Ci application to generate .wharf-ci.yml files and execute them against a kubernetes cluster",
	Long:  ``,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {

		parsedLogLevel, err := parseLevel(loglevel)
		if err != nil {
			parsedLogLevel = logger.LevelInfo
		}
		logger.AddOutput(parsedLogLevel, consolepretty.Default)

		if err != nil {
			log.Warn().WithString("loglevel", parsedLogLevel.String()).Message("Unable to parse loglevel.")
		} else {
			log.Debug().WithString("loglevel", parsedLogLevel.String()).Message("Setting log-level.")
		}

		Kubeconfig, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			log.Warn().WithError(err).Message("Failed to load kube-config")
		} else {
			log.Debug().WithString("host", Kubeconfig.Host).Message("Loaded kube-config")
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Panic().WithError(err).Message("Execution failed.")
	}
}

func init() {
	home := homeDir()
	rootCmd.PersistentFlags().StringVar(&loglevel, "loglevel", "info", "Show debug information")
	rootCmd.PersistentFlags().StringVar(&kubeconfigPath, "kubeconfig", filepath.Join(home, ".kube", "config"), "Path to kubeconfig file")
	rootCmd.PersistentFlags().StringVar(&Namespace, "namespace", "default", "Namespace to spawn resources in")
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

// parseLevel is added in https://github.com/iver-wharf/wharf-core/pull/14 but
// that PR has not yet merged at the time or writing.
func parseLevel(lvl string) (logger.Level, error) {
	switch strings.ToLower(lvl) {
	case "debug":
		return logger.LevelDebug, nil
	case "info":
		return logger.LevelInfo, nil
	case "warn":
		return logger.LevelWarn, nil
	case "error":
		return logger.LevelError, nil
	case "panic":
		return logger.LevelPanic, nil
	default:
		return logger.LevelDebug, fmt.Errorf("invalid logging level: %q", lvl)
	}
}
