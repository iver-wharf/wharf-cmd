package main

import (
	"fmt"

	"github.com/iver-wharf/wharf-cmd/internal/flagtypes"
	"github.com/iver-wharf/wharf-cmd/internal/lastbuild"
	"github.com/spf13/cobra"
)

var varsFlags = struct {
	env     string
	inputs  flagtypes.KeyValueArray
	buildID uint
}{}

var varsCmd = &cobra.Command{
	Use:   "vars",
	Short: "Commands for working with wharf-cmd's variable substitution",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Intentionally not calling parent PersistentPreRunE
		// to disable the SIGTERM signal hooks from rootCmd

		if varsFlags.buildID == 0 {
			buildID, err := lastbuild.GuessNext()
			if err != nil {
				return fmt.Errorf("get default for --build-id flag: %w", err)
			}
			varsFlags.buildID = buildID
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(varsCmd)

	addBuildIDFlag(varsCmd.PersistentFlags(), &varsFlags.buildID)
	addWharfYmlEnvFlag(varsCmd, varsCmd.PersistentFlags(), &varsFlags.env)
	addWharfYmlInputsFlag(varsCmd, varsCmd.PersistentFlags(), &varsFlags.inputs)
}
