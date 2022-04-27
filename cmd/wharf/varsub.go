package main

import (
	"github.com/iver-wharf/wharf-cmd/internal/flagtypes"
	"github.com/spf13/cobra"
)

var varsubFlags = struct {
	env     string
	showAll bool
	inputs  flagtypes.KeyValueArray
}{}

var varsubCmd = &cobra.Command{
	Use:   "varsub",
	Short: "Commands for working with wharf-cmd's variable substitution",
}

func init() {
	rootCmd.AddCommand(varsubCmd)

	varsubCmd.PersistentFlags().StringVarP(&varsubFlags.env, "environment", "e", "", "Environment selection")
	varsubCmd.RegisterFlagCompletionFunc("environment", completeWharfYmlEnv)
	varsubCmd.PersistentFlags().BoolVarP(&varsubFlags.showAll, "all", "a", false, "Show overridden variables")
	varsubCmd.PersistentFlags().VarP(&varsubFlags.inputs, "input", "i", "Inputs (--input key=value), can be set multiple times")
}
