package main

import (
	"github.com/iver-wharf/wharf-cmd/pkg/aggregator"
	"github.com/spf13/cobra"
)

func newAggregator() (aggregator.Aggregator, error) {
	restConfig, err := loadKubeconfig()
	if err != nil {
		return nil, err
	}
	return aggregator.NewK8sAggregator(
		aggregatorFlags.instanceID,
		rootConfig.K8s.Namespace,
		rootConfig.Aggregator,
		restConfig)
}

var aggregatorFlags = struct {
	instanceID string
}{
	instanceID: "local",
}

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

		return nil
	},
}

func init() {
	rootCmd.AddCommand(aggregatorCmd)

	aggregatorCmd.Flags().StringVar(&aggregatorFlags.instanceID, "instance", aggregatorFlags.instanceID, "Wharf instance ID, used to avoid collisions in Pod ownership.")
	addKubernetesFlags(aggregatorCmd.PersistentFlags())
}
