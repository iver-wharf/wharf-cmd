package cmd

import (
	"github.com/iver-wharf/wharf-cmd/pkg/aggregator"
	"github.com/spf13/cobra"
)

var aggregatorServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Starts forwarding build results from workers to the Wharf API",
	Long: `Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do
eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim
veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo
consequat.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		k8sAggregator, err := aggregator.NewK8sAggregator(aggregatorFlags.namespace, aggregatorFlags.restConfig)
		if err != nil {
			return err
		}
		return k8sAggregator.Serve()
	},
}

func init() {
	aggregatorCmd.AddCommand(aggregatorServeCmd)
}
