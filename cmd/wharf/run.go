package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/iver-wharf/wharf-cmd/internal/flagtypes"
	"github.com/iver-wharf/wharf-cmd/pkg/resultstore"
	"github.com/iver-wharf/wharf-cmd/pkg/tarstore"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml"
	"github.com/iver-wharf/wharf-cmd/pkg/worker"
	"github.com/iver-wharf/wharf-cmd/pkg/worker/workermodel"
	"github.com/iver-wharf/wharf-cmd/pkg/workerapi/workerserver"
	"github.com/spf13/cobra"
	"gopkg.in/typ.v4/slices"
)

var runFlags = struct {
	stage       string
	env         string
	serve       bool
	noGitIgnore bool
	inputs      flagtypes.KeyValueArray
	dryRun      flagtypes.DryRun
}{
	dryRun: flagtypes.DryRunNone,
}

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
		kubeconfig, err := loadKubeconfig()
		if err != nil {
			return err
		}
		currentDir, err := parseCurrentDir(slices.SafeGet(args, 0))
		if err != nil {
			return err
		}
		def, err := parseBuildDefinition(currentDir, wharfyml.Args{
			Env:    runFlags.env,
			Inputs: parseInputArgs(runFlags.inputs),
		})
		if err != nil {
			return err
		}

		// TODO: Change to build ID-based path, e.g /tmp/iver-wharf/wharf-cmd/builds/123/...
		//
		// May require setting of owner and SUID on wharf-cmd binary to access /tmp or similar.
		// e.g.: (root should not be used in prod)
		//  chown root $(which wharf-cmd) && chmod +4000 $(which wharf-cmd)
		store := resultstore.NewStore(resultstore.NewFS("./build_logs"))

		go func() {
			<-rootContext.Done()
			if err := store.Close(); err != nil {
				log.Warn().WithError(err).Message("Error closing store.")
			} else {
				log.Debug().Message("Successfully closed store.")
			}
		}()

		tarStore, err := tarstore.New(currentDir)
		if err != nil {
			return err
		}
		defer tarStore.Close()
		b, err := worker.NewK8s(rootContext, def,
			worker.K8sRunnerOptions{
				BuildOptions: worker.BuildOptions{
					StageFilter: runFlags.stage,
				},
				Config:        &rootConfig,
				CurrentDir:    currentDir,
				RestConfig:    kubeconfig,
				ResultStore:   store,
				SkipGitIgnore: runFlags.noGitIgnore,
				TarStore:      tarStore,
				VarSource:     def.VarSource,
				DryRun:        convDryRunFlag(runFlags.dryRun),
			})
		if err != nil {
			return err
		}

		ctx := rootContext
		if runFlags.serve {
			var server workerserver.Server
			ctx, server = startWorkerServerWithCancel(rootContext, store)
			defer server.Close()
		}

		log.Debug().Message("Successfully created builder.")
		log.Info().Message("Starting build.")
		res, err := b.Build(ctx)
		if err != nil {
			return err
		}

		if res.Status != workermodel.StatusSuccess && res.Status != workermodel.StatusCancelled {
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

func startWorkerServerWithCancel(ctx context.Context, store resultstore.Store) (context.Context, workerserver.Server) {
	ctx, cancel := context.WithCancel(ctx)
	server := workerserver.New(store, nil)

	go func() {
		<-ctx.Done()
		server.Close()
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

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().BoolVar(&runFlags.serve, "serve", false, "Serves build results over REST & gRPC and waits until terminated (e.g via SIGTERM)")
	runCmd.Flags().BoolVar(&runFlags.noGitIgnore, "no-gitignore", false, "Don't respect .gitignore files")
	runCmd.Flags().Var(&runFlags.dryRun, "dry-run", `Must be one of "none", "client", or "server"`)
	runCmd.RegisterFlagCompletionFunc("dry-run", flagtypes.CompleteDryRun)

	addWharfYmlStageFlag(runCmd, runCmd.Flags(), &runFlags.stage)
	addWharfYmlEnvFlag(runCmd, runCmd.Flags(), &runFlags.env)
	addWharfYmlInputsFlag(runCmd, runCmd.Flags(), &runFlags.inputs)
	addKubernetesFlags(runCmd.Flags())
}

func convDryRunFlag(dryRun flagtypes.DryRun) worker.DryRun {
	switch dryRun {
	case flagtypes.DryRunClient:
		return worker.DryRunClient
	case flagtypes.DryRunServer:
		return worker.DryRunServer
	default:
		return worker.DryRunNone
	}
}
