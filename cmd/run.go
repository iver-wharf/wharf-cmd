package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/iver-wharf/wharf-cmd/pkg/builder"
	"github.com/iver-wharf/wharf-cmd/pkg/core/containercreator"
	"github.com/iver-wharf/wharf-cmd/pkg/core/wharfyml"
	"github.com/spf13/cobra"
)

var flagRunPath string
var flagEnvironment string
var flagStage string
var flagStep string
var buildID int
var runDryRun bool

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		stepRun, err := builder.NewK8sStepRunner("build", Kubeconfig)
		if err != nil {
			return err
		}
		stageRun := builder.NewStageRunner(stepRun)
		def, err := wharfyml.Parse(flagRunPath, make(map[containercreator.BuiltinVar]string))
		if err != nil {
			return err
		}
		stage, ok := def.Stages[flagStage]
		if !ok {
			return fmt.Errorf("stage not found in .wharf-ci.yml: %q", flagStage)
		}
		res, err := stageRun.RunStage(context.Background(), stage)
		if err != nil {
			return err
		}
		if !res.Success {
			return errors.New("some step failed")
		}
		return nil
		//vars := map[containercreator.BuiltinVar]string{}
		//runner := run.NewRunner(Kubeconfig, "")
		//runner.DryRun = runDryRun
		//return runner.Run(runPath, environment, Namespace, stage, buildID,
		//	git.NewGitPropertiesMap("", "", ""), vars)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().StringVarP(&flagRunPath, "path", "p", ".wharf-ci.yml", "Path to .wharf-ci file")
	runCmd.Flags().StringVarP(&flagEnvironment, "environment", "e", "", "Environment")
	runCmd.Flags().StringVar(&flagStage, "stage", "", "Stage to run")
	runCmd.Flags().StringVar(&flagStep, "step", "", "Step to run")
	runCmd.Flags().IntVarP(&buildID, "build-id", "b", -1, "Build ID")
	runCmd.Flags().BoolVar(&runDryRun, "dry-run", false, `Only log what's planned, don't actually start any pods`)

	runCmd.MarkFlagRequired("stage")
}
