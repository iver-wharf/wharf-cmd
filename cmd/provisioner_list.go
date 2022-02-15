package cmd

import (
	"context"

	"github.com/iver-wharf/wharf-cmd/pkg/provisioner"
	"github.com/spf13/cobra"
)

var provisionerListCmd = &cobra.Command{
	Use:   "list",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := provisioner.NewK8sProvisioner("default", Kubeconfig)
		if err != nil {
			return err
		}

		workers, err := p.ListWorkers(context.Background())
		if err != nil {
			return err
		}

		log.Info().WithInt("count", len(workers)).Message("Fetched workers with matching labels.")
		for i, worker := range workers {
			log.Info().WithInt("index", i).
				WithString("workerID", string(worker.ID)).
				WithString("name", worker.Name).
				WithStringer("status", worker.Status).
				Message("")
		}

		return nil
	},
}

func init() {
	provisionerCmd.AddCommand(provisionerListCmd)
}
