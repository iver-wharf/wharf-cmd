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
	return aggregator.NewK8sAggregator(&rootConfig, restConfig)
}

var aggregatorCmd = &cobra.Command{
	Use:   "aggregator",
	Short: "Forwards build results from workers to the Wharf API.",
	Long: `Streams build results from workers to the Wharf API through gRPC.
After streaming from a worker is done, the aggregator will kill it using the
kill endpoint.`,
}

func init() {
	rootCmd.AddCommand(aggregatorCmd)

	addKubernetesFlags(aggregatorCmd.PersistentFlags())
}
