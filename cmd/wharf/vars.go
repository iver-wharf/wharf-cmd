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

	varsCmd.PersistentFlags().StringVarP(&varsFlags.env, "environment", "e", "", "Environment selection")
	varsCmd.RegisterFlagCompletionFunc("environment", completeWharfYmlEnv)
	varsCmd.PersistentFlags().BoolVarP(&varsFlags.showAll, "all", "a", false, "Show overridden variables")
	varsCmd.PersistentFlags().VarP(&varsFlags.inputs, "input", "i", "Inputs (--input key=value), can be set multiple times")
	varsCmd.RegisterFlagCompletionFunc("input", completeWharfYmlInputs)
}
