package main

import (
	"github.com/iver-wharf/wharf-cmd/pkg/provisioner"
	"github.com/spf13/cobra"
)

func newProvisioner() (provisioner.Provisioner, error) {
	restConfig, err := loadKubeconfig()
	if err != nil {
		return nil, err
	}
	return provisioner.NewK8sProvisioner(&rootConfig, restConfig)
}

var provisionerCmd = &cobra.Command{
	Use:   "provisioner",
	Short: "Provision workers that runs builds",
	Long: `Provisions workers, which are Kubernetes pods running wharf-cmd
that clones the repository and run the .wharf-ci.yml file inside the repo.

The "wharf provisioner" act as a fire-and-forget, where the entire build
orchestration is handled inside the Kubernetes cluster, in comparison to the
"wharf run" command that uses your local machine to orchestrate the build.
`,
}

func init() {
	rootCmd.AddCommand(provisionerCmd)

	addKubernetesFlags(provisionerCmd.PersistentFlags())
}
