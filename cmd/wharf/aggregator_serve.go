package main

import (
	"github.com/spf13/cobra"
)

var aggregatorServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Aggregator forwards build results from workers to the Wharf API",
	Long: `The aggregator tool is used to stream build results from workers to
the Wharf API through gRPC, killing the worker upon completion.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		k8sAggregator, err := newAggregator()
		if err != nil {
			return err
		}
		return k8sAggregator.Serve(rootContext)
	},
}

func init() {
	aggregatorCmd.AddCommand(aggregatorServeCmd)
}
