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
		p, err := provisioner.NewK8sProvisioner(provisionerFlags.namespace, provisionerFlags.restConfig)
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
			WithString("workerID", string(worker.ID)).
			Message("Created worker")

		return nil
	},
}

func init() {
	provisionerCmd.AddCommand(provisionerCreateCmd)
	provisionerCreateCmd.Flags().StringVarP(&provisionerCreateFlags.stage, "stage", "s", "", "Stage to run (will run all stages if unset)")
	provisionerCreateCmd.RegisterFlagCompletionFunc("stage", completeWharfYmlStage)
	provisionerCreateCmd.Flags().StringVarP(&provisionerCreateFlags.env, "environment", "e", "", "Environment selection")
	provisionerCreateCmd.RegisterFlagCompletionFunc("environment", completeWharfYmlEnv)
	provisionerCreateCmd.Flags().VarP(&provisionerCreateFlags.inputs, "input", "i", "Inputs (--input key=value), can be set multiple times")
	provisionerCreateCmd.RegisterFlagCompletionFunc("input", completeWharfYmlInputs)
	provisionerCreateCmd.Flags().StringVar(&provisionerCreateFlags.subdir, "subdir", "", "Subdirectory of repository where .wharf-ci.yml file is found.")
}
