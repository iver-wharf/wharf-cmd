package cmd

import (
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
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
		log.SetFormatter(&log.TextFormatter{
			FullTimestamp:             true,
			TimestampFormat:           "15:04:05",
			ForceColors:               true,
			DisableColors:             false,
			EnvironmentOverrideColors: false,
			DisableTimestamp:          false,
			DisableSorting:            false,
			SortingFunc:               nil,
			DisableLevelTruncation:    false,
			QuoteEmptyFields:          false,
			FieldMap:                  nil,
			CallerPrettyfier:          nil,
		})

		parsedLogLevel, err := log.ParseLevel(loglevel)
		if err != nil {
			log.WithField("loglevel", loglevel).Warnln("Unable to parse loglevel")
		} else {
			log.SetLevel(parsedLogLevel)
			log.WithField("loglevel", parsedLogLevel).Traceln("setting log-level")
		}

		Kubeconfig, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			log.WithError(err).Infoln("failed to load kube-config")
		} else {
			log.WithField("host", Kubeconfig.Host).Traceln("loaded kube-config")
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
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
