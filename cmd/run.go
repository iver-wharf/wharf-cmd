package cmd

import (
	"github.com/iver-wharf/wharf-cmd/pkg/builder"
	"github.com/iver-wharf/wharf-cmd/pkg/core/containercreator"
	"github.com/iver-wharf/wharf-cmd/pkg/core/wharfyml"
	"github.com/spf13/cobra"
)

var runPath string
var environment string
var stage string
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
		def, err := wharfyml.Parse(".wharf-ci.yml", make(map[containercreator.BuiltinVar]string))
		if err != nil {
			return err
		}
		return stepRun.RunStep(def.Stages["test"].Steps[0]).Error
		//vars := map[containercreator.BuiltinVar]string{}
		//runner := run.NewRunner(Kubeconfig, "")
		//runner.DryRun = runDryRun
		//return runner.Run(runPath, environment, Namespace, stage, buildID,
		//	git.NewGitPropertiesMap("", "", ""), vars)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().StringVarP(&runPath, "path", "p", ".wharf-ci.yml", "Path to .wharf-ci file")
	runCmd.Flags().StringVarP(&environment, "environment", "e", "", "Environment")
	runCmd.Flags().StringVarP(&stage, "stage", "s", "", "Stage to run")
	runCmd.Flags().IntVarP(&buildID, "build-id", "b", -1, "Build ID")
	runCmd.Flags().BoolVar(&runDryRun, "dry-run", false, `Only log what's planned, don't actually start any pods`)
}
