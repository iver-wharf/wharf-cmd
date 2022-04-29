package main

import (
	"github.com/iver-wharf/wharf-cmd/pkg/provisioner"
	"github.com/spf13/cobra"
)

var provisionerCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Starts a build via a new worker inside a Kubernetes pod",
	Long: `Creates a new Kubernetes pod that clones a Git repo and
a container running "wharf run" to perform the build.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := provisioner.NewK8sProvisioner(rootConfig.Provisioner, provisionerFlags.restConfig)
		if err != nil {
			return err
		}

		worker, err := p.CreateWorker(rootContext)
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
