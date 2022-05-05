package main

import (
	"github.com/iver-wharf/wharf-cmd/pkg/provisioner"
	"github.com/spf13/cobra"
)

var provisionerFlags = struct {
	instanceID string
}{
	instanceID: "local",
}

func newProvisioner() (provisioner.Provisioner, error) {
	restConfig, ns, err := loadKubeconfig()
	if err != nil {
		return nil, err
	}
	rootConfig.K8s.Namespace = ns
	return provisioner.NewK8sProvisioner(
		provisionerFlags.instanceID,
		rootConfig.Provisioner,
		rootConfig.K8s.Namespace,
		restConfig)
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

	provisionerCmd.Flags().StringVar(&provisionerFlags.instanceID, "instance", provisionerFlags.instanceID, "Wharf instance ID, used to avoid collisions in Pod ownership.")
	addKubernetesFlags(provisionerCmd.PersistentFlags(), &rootConfig.K8s.Namespace)
}
