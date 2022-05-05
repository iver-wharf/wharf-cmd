package main

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
	Long: `Streams build results from workers to the Wharf API through gRPC.
After streaming from a worker is done, the aggregator will kill it using the
kill endpoint.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := callParentPersistentPreRuns(cmd, args); err != nil {
			return err
		}

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
