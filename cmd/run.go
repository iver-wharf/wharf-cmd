package cmd

import (
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
		def, err := wharfyml.Parse(flagRunPath, make(map[containercreator.BuiltinVar]string))
		if err != nil {
			return err
		}
		stage, ok := def.Stages[flagStage]
		if !ok {
			return fmt.Errorf("stage not found in .wharf-ci.yml: %q", flagStage)
		}
		var step *wharfyml.Step
		for _, s := range stage.Steps {
			if s.Name == flagStep {
				step = &s
				break
			}
		}
		if step == nil {
			return fmt.Errorf("step in stage %q not found in .wharf-ci.yml: %q", flagStage, flagStep)
		}
		return stepRun.RunStep(*step).Error
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
	runCmd.MarkFlagRequired("step")
}
