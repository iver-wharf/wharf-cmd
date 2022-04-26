package main

import (
	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var provisionerFlags = struct {
	k8sOverrides clientcmd.ConfigOverrides

	restConfig *rest.Config
	namespace  string
}{}

var provisionerCmd = &cobra.Command{
	Use:   "provisioner",
	Short: "Provision workers that runs builds",
	Long: `Provisions workers, which are Kubernetes pods running wharf-cmd
that clones the repository and run the .wharf-ci.yml file inside the repo.

The "wharf provisioner" act as a fire-and-forget, where the entire build
orchestration is handled inside the Kubernetes cluster, in comparison to the
"wharf run" command that uses your local machine to orchestrate the build.
`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		go handleCancelSignals(rootCancel)
		restConfig, ns, err := loadKubeconfig(provisionerFlags.k8sOverrides)
		if err != nil {
			return err
		}
		provisionerFlags.restConfig = restConfig
		provisionerFlags.namespace = ns
		return nil
	},
}

func init() {
	rootCmd.AddCommand(provisionerCmd)

	addKubernetesFlags(provisionerCmd.PersistentFlags(), &provisionerFlags.k8sOverrides)
}
