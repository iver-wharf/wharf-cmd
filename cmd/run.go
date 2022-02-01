package cmd

import (
	"github.com/iver-wharf/wharf-cmd/pkg/core/containercreator"
	"github.com/iver-wharf/wharf-cmd/pkg/core/containercreator/git"
	"github.com/iver-wharf/wharf-cmd/pkg/run"
	"github.com/spf13/cobra"
)

var runPath string
var environment string
var stage string
var buildID int

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		vars := map[containercreator.BuiltinVar]string{}
		return run.NewRunner(Kubeconfig, "").Run(runPath, environment, Namespace, stage, buildID,
			git.NewGitPropertiesMap("", "", ""), vars)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().StringVarP(&runPath, "path", "p", ".wharf-ci.yml", "Path to .wharf-ci file")
	runCmd.Flags().StringVarP(&environment, "environment", "e", "", "Environment")
	runCmd.Flags().StringVarP(&stage, "stage", "s", "", "Stage to run")
	runCmd.Flags().IntVarP(&buildID, "build-id", "b", -1, "Build ID")
}
