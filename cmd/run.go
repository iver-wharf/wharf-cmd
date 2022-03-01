package cmd

import (
	"context"
	"errors"
	"time"

	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml"
	"github.com/iver-wharf/wharf-cmd/pkg/worker"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

var runFlags = struct {
	path         string
	stage        string
	k8sOverrides clientcmd.ConfigOverrides
}{}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		kubeconfig, ns, err := loadKubeconfig(runFlags.k8sOverrides)
		if err != nil {
			return err
		}
		stepRun, err := worker.NewK8sStepRunner(ns, kubeconfig)
		if err != nil {
			return err
		}
		stageRun := worker.NewStageRunner(stepRun)
		b := worker.New(stageRun)
		def, errs := wharfyml.ParseFile(runFlags.path)
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
			StageFilter: runFlags.stage,
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

	runCmd.Flags().StringVarP(&runFlags.path, "path", "p", ".wharf-ci.yml", "Path to .wharf-ci file")
	runCmd.Flags().StringVarP(&runFlags.stage, "stage", "s", "", "Stage to run (will run all stages if unset)")
	addKubernetesFlags(runCmd.Flags(), &runFlags.k8sOverrides)
}
