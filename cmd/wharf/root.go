package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

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

	cancelGracePeriod = 10 * time.Second
)

var isLoggingInitialized bool
var loglevel string

var rootCmd = &cobra.Command{
	SilenceErrors: true,
	SilenceUsage:  true,
	Use:           "wharf",
	Short:         "Ci application to generate .wharf-ci.yml files and execute them against a kubernetes cluster",
	Long:          ``,
}

var rootContext, rootCancel = context.WithCancel(context.Background())

func addKubernetesFlags(flagSet *pflag.FlagSet, overrides *clientcmd.ConfigOverrides) {
	overrideFlags := clientcmd.RecommendedConfigOverrideFlags("k8s-")
	clientcmd.BindOverrideFlags(overrides, flagSet, overrideFlags)
}

func loadKubeconfig(overrides clientcmd.ConfigOverrides) (*rest.Config, string, error) {
	loader := clientcmd.NewDefaultClientConfigLoadingRules()
	clientConf := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loader, &overrides)
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
	rootCmd.PersistentFlags().StringVar(&loglevel, "loglevel", "info", "Show debug information")
}

func initLoggingIfNeeded() {
	if !isLoggingInitialized {
		initLogging()
	}
}

func initLogging() {
	parsedLogLevel, err := parseLevel(loglevel)
	if err != nil {
		parsedLogLevel = logger.LevelInfo
	}
	logConfig := consolepretty.DefaultConfig
	if parsedLogLevel != logger.LevelDebug {
		logConfig.DisableCaller = true
		logConfig.DisableDate = true
		logConfig.ScopeMinLengthAuto = false
	}
	logger.AddOutput(parsedLogLevel, consolepretty.New(logConfig))
	if err != nil {
		log.Warn().WithStringer("loglevel", parsedLogLevel).Message("Unable to parse loglevel. Defaulting to 'INFO'.")
		parsedLogLevel = logger.LevelInfo
	} else {
		log.Debug().WithStringer("loglevel", parsedLogLevel).Message("Setting log-level.")
	}
	isLoggingInitialized = true
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

func handleCancelSignals(cancel context.CancelFunc) {
	waitForCancelSignal()
	log.Info().WithDuration("gracePeriod", cancelGracePeriod).Message("Cancelling build. Press ^C again to force quit.")
	go func() {
		waitForCancelSignal()
		log.Warn().Message("Received second interrupt. Force quitting now.")
		os.Exit(exitCodeCancelForceQuit)
	}()
	go func() {
		time.AfterFunc(cancelGracePeriod, func() {
			log.Warn().Message("Failed to cancel within grace period. Force quitting now.")
			os.Exit(exitCodeCancelTimeout)
		})
	}()
	cancel()
}

func waitForCancelSignal() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGHUP)
	<-ch
}
