package cmd

import (
	"context"

	"github.com/iver-wharf/wharf-cmd/pkg/provisioner"
	"github.com/spf13/cobra"
)

var deleteWorkerID string
var provisionerDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := provisioner.NewK8sProvisioner("default", Kubeconfig)
		if err != nil {
			return err
		}

		if err = p.DeleteWorker(context.Background(), deleteWorkerID); err == nil {
			log.Info().Message("Successfully deleted worker.")
		}

		return err
	},
}

func init() {
	provisionerDeleteCmd.PersistentFlags().StringVar(&deleteWorkerID, "id", "", "ID of the worker to delete")
}
