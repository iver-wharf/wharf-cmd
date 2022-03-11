package cmd

import (
	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var aggregatorFlags = struct {
	k8sOverrides clientcmd.ConfigOverrides

	restConfig *rest.Config
	namespace  string
}{}

var aggregatorCmd = &cobra.Command{
	Use:   "aggregator",
	Short: "Forwards build results from workers to the Wharf API.",
	Long: `Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do
eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim
veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo
consequat.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		restConfig, ns, err := loadKubeconfig(aggregatorFlags.k8sOverrides)
		if err != nil {
			return err
		}
		aggregatorFlags.restConfig = restConfig
		aggregatorFlags.namespace = ns
		return nil
	},
}

func init() {
	rootCmd.AddCommand(aggregatorCmd)

	addKubernetesFlags(aggregatorCmd.PersistentFlags(), &aggregatorFlags.k8sOverrides)
}
