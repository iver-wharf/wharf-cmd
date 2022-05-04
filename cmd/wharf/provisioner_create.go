package main

import (
	"github.com/iver-wharf/wharf-cmd/internal/flagtypes"
	"github.com/iver-wharf/wharf-cmd/pkg/provisioner"
	"github.com/spf13/cobra"
)

var provisionerCreateFlags = struct {
	stage  string
	env    string
	subdir string
	inputs flagtypes.KeyValueArray
}{}

var provisionerCreateCmd = &cobra.Command{
	Use:   "create <repo>",
	Short: "Starts a build via a new worker inside a Kubernetes pod",
	Long: `Creates a new Kubernetes pod that clones a Git repo and
a container running "wharf run" to perform the build.

The <repo> argument is used by Git to clone the repository, such as:

  wharf provisioner create https://github.com/iver-wharf/wharf-cmd
  wharf provisioner create ssh://git@github.com/iver-wharf/wharf-cmd.git`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := newProvisioner()
		if err != nil {
			return err
		}

		worker, err := p.CreateWorker(rootContext, provisioner.WorkerArgs{
			GitCloneURL: args[0],
			SubDir:      provisionerCreateFlags.subdir,
			Inputs:      parseInputArgs(provisionerCreateFlags.inputs),
			Environment: provisionerCreateFlags.env,
			Stage:       provisionerCreateFlags.stage,
		})
		if err != nil {
			return err
		}

		log.Info().WithString("name", worker.Name).
			WithString("workerID", string(worker.WorkerID)).
			Message("Created worker")

		return nil
	},
}

func init() {
	provisionerCmd.AddCommand(provisionerCreateCmd)

	provisionerCreateCmd.Flags().StringVar(&provisionerCreateFlags.subdir, "subdir", "", "Subdirectory of repository where .wharf-ci.yml file is found.")

	addWharfYmlStageFlag(provisionerCreateCmd, &provisionerCreateFlags.stage)
	addWharfYmlEnvFlag(provisionerCreateCmd, &provisionerCreateFlags.env)
	addWharfYmlInputsFlag(provisionerCreateCmd, &provisionerCreateFlags.inputs)
}
