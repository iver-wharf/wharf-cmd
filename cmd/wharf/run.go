package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/iver-wharf/wharf-cmd/internal/gitutil"
	"github.com/iver-wharf/wharf-cmd/pkg/resultstore"
	"github.com/iver-wharf/wharf-cmd/pkg/varsub"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml"
	"github.com/iver-wharf/wharf-cmd/pkg/worker"
	"github.com/iver-wharf/wharf-cmd/pkg/worker/workermodel"
	"github.com/iver-wharf/wharf-cmd/pkg/workerapi/workerserver"
	"github.com/spf13/cobra"
	"gopkg.in/typ.v3/pkg/slices"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	cancelGracePeriod = 10 * time.Second
)

var runFlags = struct {
	stage        string
	env          string
	serve        bool
	k8sOverrides clientcmd.ConfigOverrides
	noGitIgnore  bool
}{}

var runCmd = &cobra.Command{
	Use:   "run [path]",
	Short: "Runs a new build from a .wharf-ci.yml file",
	Long: `Runs a new build in a Kubernetes cluster using pods
based on a .wharf-ci.yml file.

Use the optional "path" argument to specify a .wharf-ci.yml file or a
directory containing a .wharf-ci.yml file. Defaults to current directory ("./")

If no stage is specified via --stage then wharf will run all stages
in sequence, based on their order of declaration in the .wharf-ci.yml file.

All steps in each stage will be run in parallel for each stage.

Read more about the .wharf-ci.yml file here:
https://iver-wharf.github.io/#/usage-wharfyml/`,
	Args: cobra.MaximumNArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"yml"}, cobra.ShellCompDirectiveFilterFileExt
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		kubeconfig, ns, err := loadKubeconfig(runFlags.k8sOverrides)
		if err != nil {
			return err
		}
		currentDir, err := parseCurrentDir(slices.SafeGet(args, 0))
		if err != nil {
			return err
		}
		def, err := parseBuildDefinition(currentDir)
		if err != nil {
			return err
		}
		log.Debug().Message("Successfully parsed .wharf-ci.yml")

		// TODO: Change to build ID-based path, e.g /tmp/iver-wharf/wharf-cmd/builds/123/...
		//
		// May require setting of owner and SUID on wharf-cmd binary to access /tmp or similar.
		// e.g.: (root should not be used in prod)
		//  chown root $(which wharf-cmd) && chmod +4000 $(which wharf-cmd)
		store := resultstore.NewStore(resultstore.NewFS("./build_logs"))
		b, err := worker.NewK8s(context.Background(), def,
			worker.K8sRunnerOptions{
				Namespace:  ns,
				RestConfig: kubeconfig,
				Store:      store,
				BuildOptions: worker.BuildOptions{
					StageFilter: runFlags.stage,
					RepoDir:     currentDir,
				},
				SkipGitIgnore: runFlags.noGitIgnore,
			})
		if err != nil {
			return err
		}

		ctx, cancel := context.WithCancel(context.Background())
		if runFlags.serve {
			var server workerserver.Server
			ctx, server = startWorkerServerWithCancel(ctx, store)
			defer server.Close()
		}

		go handleCancelSignals(func() {
			go func() {
				time.AfterFunc(cancelGracePeriod, func() {
					log.Warn().Message("Failed to cancel within grace period. Force quitting now.")
					os.Exit(3)
				})
			}()
			cancel()
		})

		log.Debug().Message("Successfully created builder.")
		log.Info().Message("Starting build.")
		res, err := b.Build(ctx)
		if err != nil {
			return err
		}
		if res.Status != workermodel.StatusSuccess {
			return errors.New("build failed")
		}
		log.Info().
			WithDuration("dur", res.Duration.Truncate(time.Second)).
			WithStringer("status", res.Status).
			Message("Done with build.")

		if err := store.Freeze(); err != nil {
			return fmt.Errorf("freeze result store: %w", err)
		}
		if runFlags.serve {
			<-ctx.Done()
		}
		return nil
	},
}

func parseCurrentDir(dirArg string) (string, error) {
	if dirArg == "" {
		return os.Getwd()
	}
	abs, err := filepath.Abs(dirArg)
	if err != nil {
		return "", err
	}
	stat, err := os.Stat(abs)
	if err != nil {
		return "", err
	}
	if !stat.IsDir() {
		dir, file := filepath.Split(abs)
		if file == ".wharf-ci.yml" {
			return dir, nil
		}
		return "", fmt.Errorf("path is neither a dir nor a .wharf-ci.yml file: %s", abs)
	}
	stat, err = os.Stat(filepath.Join(abs, ".wharf-ci.yml"))
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("missing .wharf-ci.yml file in dir: %s", abs)
		}
		return "", err
	}
	return abs, nil
}

func parseBuildDefinition(currentDir string) (wharfyml.Definition, error) {
	var varSources varsub.SourceSlice

	varFileSource, errs := wharfyml.ParseVarFiles(currentDir)
	if len(errs) > 0 {
		logParseErrors(errs, currentDir)
		return wharfyml.Definition{}, errors.New("failed to parse variable files")
	}
	if varFileSource != nil {
		varSources = append(varSources, varFileSource)
	}

	gitStats, err := gitutil.StatsFromExec(currentDir)
	if err != nil {
		log.Warn().WithError(err).
			Message("Failed to get REPO_ and GIT_ variables from Git. Skipping those.")
	} else {
		log.Debug().Message("Read REPO_ and GIT_ variables from Git:\n" +
			gitStats.String())
		varSources = append(varSources, gitStats)
	}

	ymlPath := filepath.Join(currentDir, ".wharf-ci.yml")
	log.Debug().WithString("path", ymlPath).Message("Parsing .wharf-ci.yml file.")
	def, errs := wharfyml.ParseFile(ymlPath, wharfyml.Args{
		Env:       runFlags.env,
		VarSource: varSources,
	})
	if len(errs) > 0 {
		logParseErrors(errs, currentDir)
		return wharfyml.Definition{}, errors.New("failed to parse .wharf-ci.yml")
	}
	return def, nil
}

func startWorkerServerWithCancel(ctx context.Context, store resultstore.Store) (context.Context, workerserver.Server) {
	ctx, cancel := context.WithCancel(ctx)
	server := workerserver.New(store, nil)

	go func() {
		select {
		case <-ctx.Done():
			server.Close()
		}
	}()

	go func() {
		const address = "0.0.0.0:5010"
		log.Info().WithString("address", address).
			Message("Serving build results via REST & gRPC.")
		defer cancel()
		if err := server.Serve(address); err != nil {
			log.Error().WithError(err).Message("Server error.")
		}
	}()

	return ctx, server
}

func handleCancelSignals(f func()) {
	waitForCancelSignal()
	log.Info().WithDuration("gracePeriod", cancelGracePeriod).Message("Cancelling build. Press ^C again to force quit.")
	go func() {
		waitForCancelSignal()
		log.Warn().Message("Received second interrupt. Force quitting now.")
		os.Exit(2)
	}()
	f()
}

func waitForCancelSignal() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGHUP)
	_ = <-ch
}

func logParseErrors(errs wharfyml.Errors, currentDir string) {
	log.Warn().WithInt("errors", len(errs)).Message("Cannot run build due to parsing errors.")
	for _, err := range errs {
		var posErr wharfyml.PosError
		if errors.As(err, &posErr) {
			log.Warn().Messagef("%4d:%-4d%s",
				posErr.Source.Line, posErr.Source.Column, err.Error())
		} else {
			log.Warn().Messagef("   -:-   %s", err.Error())
		}
	}

	containsMissingBuiltin := false
	for _, err := range errs {
		if errors.Is(err, wharfyml.ErrMissingBuiltinVar) {
			containsMissingBuiltin = true
			break
		}
	}
	if containsMissingBuiltin {
		varFiles := wharfyml.ListPossibleVarsFiles(currentDir)
		var sb strings.Builder
		sb.WriteString(`Tip: You can add built-in variables to Wharf in multiple ways.

Wharf look for values in the following files:`)
		for _, file := range varFiles {
			if file.IsRel {
				continue
			}
			sb.WriteString("\n  ")
			sb.WriteString(file.PrettyPath(currentDir))
		}
		sb.WriteString(`

Wharf also looks for:
  - All ".wharf-vars.yml" in this directory or any parent directory.
  - Local Git repository and extracts GIT_ and REPO_ variables from it.
  - Environment variables.

Sample file content:
  # .wharf-vars.yml
  vars:
    REG_URL: http://harbor.example.com
`)
		log.Info().Message(sb.String())
	}
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().StringVarP(&runFlags.stage, "stage", "s", "", "Stage to run (will run all stages if unset)")
	runCmd.Flags().StringVarP(&runFlags.env, "environment", "e", "", "Environment selection")
	runCmd.Flags().BoolVar(&runFlags.serve, "serve", false, "Serves build results over REST & gRPC and waits until terminated (e.g via SIGTERM)")
	runCmd.Flags().BoolVar(&runFlags.noGitIgnore, "no-gitignore", false, "Don't respect .gitignore files")
	addKubernetesFlags(runCmd.Flags(), &runFlags.k8sOverrides)
}
