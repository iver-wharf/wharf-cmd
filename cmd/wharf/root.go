package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/iver-wharf/wharf-cmd/internal/flagtypes"
	"github.com/iver-wharf/wharf-cmd/pkg/config"
	"github.com/iver-wharf/wharf-core/v2/pkg/app"
	"github.com/iver-wharf/wharf-core/v2/pkg/logger"
	"github.com/iver-wharf/wharf-core/v2/pkg/logger/consolepretty"
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

var (
	isLoggingInitialized bool

	rootContext, rootCancel = context.WithCancel(context.Background())

	k8sOverridesFlags clientcmd.ConfigOverrides

	rootConfig     config.Config
	runAfterConfig []func()

	toCloseBeforeForceQuit []io.Closer
)

var rootFlags = struct {
	loglevel flagtypes.LogLevel
}{
	loglevel: flagtypes.LogLevel(logger.LevelInfo),
}

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

func addKubernetesFlags(flagSet *pflag.FlagSet) {
	runAfterConfig = append(runAfterConfig, func() {
		overrideFlags := clientcmd.RecommendedConfigOverrideFlags("k8s-")
		overrideFlags.CurrentContext.Default = rootConfig.K8s.Context
		overrideFlags.ContextOverrideFlags.Namespace.Default = rootConfig.K8s.Namespace
		clientcmd.BindOverrideFlags(&k8sOverridesFlags, flagSet, overrideFlags)
	})
}

func loadKubeconfig() (*rest.Config, error) {
	loader := clientcmd.NewDefaultClientConfigLoadingRules()
	clientConf := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loader, &k8sOverridesFlags)
	restConf, err := clientConf.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("load kubeconfig: %w", err)
	}
	rootConfig.K8s.Namespace, _, err = clientConf.Namespace()
	if err != nil {
		return nil, fmt.Errorf("get namespace to use: %w", err)
	}
	log.Debug().
		WithString("namespace", rootConfig.K8s.Namespace).
		WithString("host", restConf.Host).
		Message("Loaded kube-config")
	return restConf, nil
}

func execute(version app.Version) {
	var err error
	if rootConfig, err = config.LoadConfig(); err != nil {
		initLoggingIfNeeded()
		log.Error().Messagef("Config load: %s", err)
		os.Exit(exitCodeLoadConfigError)
	}

	// Some code we want to run AFTER all init()'s and AFTER wharf-cmd-config.yml
	// has been loaded; but BEFORE cobra starts parsing all arguments as flags.
	for _, f := range runAfterConfig {
		f()
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
	rootCmd.PersistentFlags().VarP(&rootFlags.loglevel, "loglevel", "l", "Show debug information")
	rootCmd.RegisterFlagCompletionFunc("loglevel", flagtypes.CompleteLogLevel)
	runAfterConfig = append(runAfterConfig, func() {
		rootCmd.PersistentFlags().StringVar(&rootConfig.InstanceID, "instance", rootConfig.InstanceID, "Wharf instance ID, used to avoid collisions in Pod ownership.")
	})
}

func initLoggingIfNeeded() {
	if !isLoggingInitialized {
		initLogging()
	}
}

func initLogging() {
	logConfig := consolepretty.DefaultConfig
	if rootFlags.loglevel.Level() != logger.LevelDebug {
		logConfig.DisableCaller = true
		logConfig.DisableDate = true
		logConfig.ScopeMinLengthAuto = false
	} else {
		logConfig.ScopeMaxLength = 16
	}
	logger.AddOutput(rootFlags.loglevel.Level(), consolepretty.New(logConfig))
	isLoggingInitialized = true
}

func handleCancelSignals(cancel context.CancelFunc) {
	ch := newCancelSignalChan()
	<-ch
	log.Info().WithDuration("gracePeriod", cancelGracePeriod).
		Message("Cancelling build. Press ^C again to force quit.")
	cancel()

	select {
	case <-ch:
		log.Warn().Message("Received second interrupt. Force quitting now.")
		forceQuit(exitCodeCancelForceQuit)
	case <-time.After(cancelGracePeriod):
		log.Warn().Message("Failed to cancel within grace period. Force quitting now.")
		forceQuit(exitCodeCancelTimeout)
	}
}

func forceQuit(exitCode int) {
	for _, closer := range toCloseBeforeForceQuit {
		closer.Close()
	}
	os.Exit(exitCode)
}

func closeBeforeForceQuit(closer io.Closer) {
	toCloseBeforeForceQuit = append(toCloseBeforeForceQuit, closer)
}

func newCancelSignalChan() <-chan os.Signal {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	return ch
}
