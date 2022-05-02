package main

import (
	"github.com/spf13/cobra"
)

var provisionerListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists current workers inside Kubernetes",
	Long: `Lists wharf worker pods inside Kubernetes
that are either scheduling, running, or completed.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := newProvisioner()
		if err != nil {
			return err
		}

		workers, err := p.ListWorkers(rootContext)
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
