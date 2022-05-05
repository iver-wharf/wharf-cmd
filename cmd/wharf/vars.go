package main

import (
	"github.com/iver-wharf/wharf-cmd/internal/flagtypes"
	"github.com/spf13/cobra"
)

var varsFlags = struct {
	env    string
	inputs flagtypes.KeyValueArray
}{}

var varsCmd = &cobra.Command{
	Use:   "vars",
	Short: "Commands for working with wharf-cmd's variable substitution",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Intentionally empty, to disable the SIGTERM signal hooks from rootCmd
		return nil
	},
}

func init() {
	rootCmd.AddCommand(varsCmd)

	addWharfYmlEnvFlag(varsCmd, varsCmd.PersistentFlags(), &varsFlags.env)
	addWharfYmlInputsFlag(varsCmd, varsCmd.PersistentFlags(), &varsFlags.inputs)
}
