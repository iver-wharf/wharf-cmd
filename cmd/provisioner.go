package cmd

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
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
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
