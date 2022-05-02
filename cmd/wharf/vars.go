package main

import (
	"github.com/iver-wharf/wharf-cmd/internal/flagtypes"
	"github.com/spf13/cobra"
)

var varsFlags = struct {
	env     string
	showAll bool
	inputs  flagtypes.KeyValueArray
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

	varsCmd.PersistentFlags().BoolVarP(&varsFlags.showAll, "all", "a", false, "Show overridden variables")

	addWharfYmlEnvFlag(varsCmd, &varsFlags.env)
	addWharfYmlInputsFlag(varsCmd, &varsFlags.inputs)
}
