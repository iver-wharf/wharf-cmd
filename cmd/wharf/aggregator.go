package main

import (
	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
)

var aggregatorFlags = struct {
	restConfig *rest.Config
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

		restConfig, ns, err := loadKubeconfig()
		if err != nil {
			return err
		}
		aggregatorFlags.restConfig = restConfig
		rootConfig.K8s.Namespace = ns
		return nil
	},
}

func init() {
	rootCmd.AddCommand(aggregatorCmd)

	addKubernetesFlags(aggregatorCmd.PersistentFlags(), &rootConfig.K8s.Namespace)
}
