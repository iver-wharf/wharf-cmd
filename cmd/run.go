package cmd

import (
	"context"
	"errors"
	"time"

	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml"
	"github.com/iver-wharf/wharf-cmd/pkg/worker"
	"github.com/spf13/cobra"
)

var flagRunPath string
var flagStage string

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
		stepRun, err := worker.NewK8sStepRunner("build", Kubeconfig)
		if err != nil {
			return err
		}
		stageRun := worker.NewStageRunner(stepRun)
		b := worker.New(stageRun)
		def, errs := wharfyml.ParseFile(flagRunPath)
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
		res, err := b.Build(context.Background(), def, worker.BuildOptions{
			StageFilter: flagStage,
		})
		if err != nil {
			return err
		}
		log.Info().
			WithDuration("dur", res.Duration.Truncate(time.Second)).
			WithStringer("status", res.Status).
			Message("Done with build.")
		if res.Status != worker.StatusSuccess {
			return errors.New("build failed")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().StringVarP(&flagRunPath, "path", "p", ".wharf-ci.yml", "Path to .wharf-ci file")
	runCmd.Flags().StringVarP(&flagStage, "stage", "s", "", "Stage to run (will run all stages if unset)")
}
