package cmd

import (
	"context"
	"errors"
	"time"

	"github.com/iver-wharf/wharf-cmd/pkg/resultstore"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml"
	"github.com/iver-wharf/wharf-cmd/pkg/worker"
	"github.com/iver-wharf/wharf-cmd/pkg/worker/workermodel"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

var runFlags = struct {
	path         string
	stage        string
	env          string
	k8sOverrides clientcmd.ConfigOverrides
}{}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Runs a new build from a .wharf-ci.yml file",
	Long: `Runs a new build in a Kubernetes cluster using pods
based on a .wharf-ci.yml file.

If no stage is specified via --stage then wharf-cmd will run all stages
in sequence, based on their order of declaration in the .wharf-ci.yml file.

All steps in each stage will be run in parallel for each stage.

Read more about the .wharf-ci.yml file here:
https://iver-wharf.github.io/#/usage-wharfyml/
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		kubeconfig, ns, err := loadKubeconfig(runFlags.k8sOverrides)
		if err != nil {
			return err
		}
		def, errs := wharfyml.ParseFile(runFlags.path, wharfyml.Args{
			Env: runFlags.env,
		})
		if len(errs) > 0 {
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
			return errors.New("failed to parse .wharf-ci.yml")
		}
		log.Debug().Message("Successfully parsed .wharf-ci.yml")
		// TODO: Change to build ID-based path, e.g /tmp/iver-wharf/wharf-cmd/builds/123/...
		//
		// May require setting of owner and SUID on wharf-cmd binary to access /var/log.
		// e.g.:
		//  chown root $(which wharf-cmd) && chmod +4000 $(which wharf-cmd)
		store := resultstore.NewStore(resultstore.NewFS("/var/log/build_logs"))
		b, err := worker.NewK8s(context.Background(), def, ns, kubeconfig, store, worker.BuildOptions{
			StageFilter: runFlags.stage,
		})
		if err != nil {
			return err
		}
		log.Debug().Message("Successfully created builder.")
		log.Info().Message("Starting build.")
		res, err := b.Build(context.Background())
		if err != nil {
			return err
		}
		log.Info().
			WithDuration("dur", res.Duration.Truncate(time.Second)).
			WithStringer("status", res.Status).
			Message("Done with build.")
		if res.Status != workermodel.StatusSuccess {
			return errors.New("build failed")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().StringVarP(&runFlags.path, "path", "p", ".wharf-ci.yml", "Path to .wharf-ci file")
	runCmd.Flags().StringVarP(&runFlags.stage, "stage", "s", "", "Stage to run (will run all stages if unset)")
	runCmd.Flags().StringVarP(&runFlags.env, "environment", "e", "", "Environment selection")
	addKubernetesFlags(runCmd.Flags(), &runFlags.k8sOverrides)
}
