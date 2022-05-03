package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/iver-wharf/wharf-cmd/internal/flagtypes"
	"github.com/iver-wharf/wharf-cmd/pkg/config"
	"github.com/iver-wharf/wharf-core/pkg/app"
	"github.com/iver-wharf/wharf-core/pkg/logger"
	"github.com/iver-wharf/wharf-core/pkg/logger/consolepretty"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	exitCodeError           = 1
	exitCodeCancelForceQuit = 2
	exitCodeCancelTimeout   = 3
	exitCodeLoadConfigError = 4

	cancelGracePeriod = 10 * time.Second
)

var isLoggingInitialized bool
var loglevel flagtypes.LogLevel = flagtypes.LogLevel(logger.LevelInfo)

var rootCmd = &cobra.Command{
	SilenceErrors: true,
	SilenceUsage:  true,
	Use:           "wharf",
	Short:         "Ci application to generate .wharf-ci.yml files and execute them against a kubernetes cluster",
	Long:          ``,
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		go handleCancelSignals(rootCancel)
	},
}

var rootContext, rootCancel = context.WithCancel(context.Background())
var rootConfig config.Config

var k8sOverridesFlags clientcmd.ConfigOverrides

func addKubernetesFlags(flagSet *pflag.FlagSet, defaultNamespace string) {
	overrideFlags := clientcmd.RecommendedConfigOverrideFlags("k8s-")
	k8sOverridesFlags.Context.Namespace = defaultNamespace
	clientcmd.BindOverrideFlags(&k8sOverridesFlags, flagSet, overrideFlags)
}

func loadKubeconfig() (*rest.Config, string, error) {
	loader := clientcmd.NewDefaultClientConfigLoadingRules()
	clientConf := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loader, &k8sOverridesFlags)
	restConf, err := clientConf.ClientConfig()
	if err != nil {
		return nil, "", fmt.Errorf("load kubeconfig: %w", err)
	}
	ns, _, err := clientConf.Namespace()
	if err != nil {
		return nil, "", fmt.Errorf("get namespace to use: %w", err)
	}
	log.Debug().
		WithString("namespace", ns).
		WithString("host", restConf.Host).
		Message("Loaded kube-config")
	return restConf, ns, nil
}

func execute(version app.Version) {
	var err error
	if rootConfig, err = config.LoadConfig(); err != nil {
		log.Error().Message(fmt.Sprintf("Config load: %s", err.Error()))
		os.Exit(exitCodeLoadConfigError)
	}

	rootCmd.Version = versionString(version)
	if err := rootCmd.Execute(); err != nil {
		initLoggingIfNeeded()
		log.Error().Message(err.Error())
		os.Exit(exitCodeError)
	}
}

func versionString(v app.Version) string {
	var sb strings.Builder
	if v.Version != "" {
		sb.WriteString(v.Version)
	} else {
		sb.WriteString("v0.0.0")
	}
	if v.BuildRef != 0 {
		fmt.Fprintf(&sb, " #%d", v.BuildRef)
	}
	if v.BuildGitCommit != "" && v.BuildGitCommit != "HEAD" {
		fmt.Fprintf(&sb, " (%s)", v.BuildGitCommit)
	}
	if v.BuildDate != (time.Time{}) {
		sb.WriteString(" built ")
		sb.WriteString(v.BuildDate.Format(time.RFC1123))
	}
	return sb.String()
}

func init() {
	cobra.OnInitialize(initLogging)
	rootCmd.InitDefaultVersionFlag()
	rootCmd.PersistentFlags().VarP(&loglevel, "loglevel", "l", "Show debug information")
	rootCmd.RegisterFlagCompletionFunc("loglevel", flagtypes.CompleteLogLevel)
}

func initLoggingIfNeeded() {
	if !isLoggingInitialized {
		initLogging()
	}
}

func initLogging() {
	logConfig := consolepretty.DefaultConfig
	if loglevel.Level() != logger.LevelDebug {
		logConfig.DisableCaller = true
		logConfig.DisableDate = true
		logConfig.ScopeMinLengthAuto = false
	}
	logger.AddOutput(loglevel.Level(), consolepretty.New(logConfig))
	isLoggingInitialized = true
}

func callParentPersistentPreRuns(cmd *cobra.Command, args []string) error {
	for {
		cmd = cmd.Parent()
		if cmd == nil {
			return nil
		}
		// Call first one we find.
		// It'll be responsible for calling its parent PersistentPreRunE
		if cmd.PersistentPreRunE != nil {
			return cmd.PersistentPreRunE(cmd, args)
		}
	}
}

func handleCancelSignals(cancel context.CancelFunc) {
	ch := newCancelSignalChan()
	<-ch
	log.Info().WithDuration("gracePeriod", cancelGracePeriod).Message("Cancelling build. Press ^C again to force quit.")
	cancel()

	select {
	case <-ch:
		log.Warn().Message("Received second interrupt. Force quitting now.")
		os.Exit(exitCodeCancelForceQuit)
	case <-time.After(cancelGracePeriod):
		log.Warn().Message("Failed to cancel within grace period. Force quitting now.")
		os.Exit(exitCodeCancelTimeout)
	}
}

func newCancelSignalChan() <-chan os.Signal {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGHUP)
	return ch
}
