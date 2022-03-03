package cmd

import (
	"github.com/spf13/cobra"
)

var provisionerCmd = &cobra.Command{
	Use:   "provisioner",
	Short: "Provision workers that runs builds",
	Long: `Provisions workers, which are Kubernetes pods running wharf-cmd
that clones the repository and run the .wharf-ci.yml file inside the repo.

The "wharf-cmd provisioner" act as a fire-and-forget, where the entire build
orchestration is handled inside the Kubernetes cluster, in comparison to the
"wharf-cmd run" command that uses your local machine to orchestrate the build.
`,
}

func init() {
	rootCmd.AddCommand(provisionerCmd)
}
