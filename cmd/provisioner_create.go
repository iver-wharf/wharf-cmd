package cmd

import (
	"context"

	"github.com/iver-wharf/wharf-cmd/pkg/provisioner"
	"github.com/spf13/cobra"
)

var provisionerCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := provisioner.NewK8sProvisioner("default", Kubeconfig)
		if err != nil {
			return err
		}

		worker, err := p.CreateWorker(context.Background())
		if err != nil {
			return err
		}

		log.Info().WithString("name", worker.Name).
			WithString("workerID", string(worker.ID)).
			Message("Created worker")

		return nil
	},
}

func init() {
	provisionerCmd.AddCommand(provisionerCreateCmd)
}
