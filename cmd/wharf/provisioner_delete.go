package main

import (
	"github.com/spf13/cobra"
)

var deleteWorkerID string
var provisionerDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Terminates a worker in Kubernetes",
	Long: `Terminates a wharf worker pod in Kubernetes, effectively
cancelling the build.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := newProvisioner()
		if err != nil {
			return err
		}

		if err = p.DeleteWorker(rootContext, deleteWorkerID); err == nil {
			log.Info().Message("Successfully deleted worker.")
		}

		return err
	},
}

func init() {
	provisionerDeleteCmd.Flags().StringVar(&deleteWorkerID, "id", "", "ID of the worker to delete")
	provisionerCmd.AddCommand(provisionerDeleteCmd)
}
