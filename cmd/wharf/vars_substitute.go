package main

import (
	"os"

	"github.com/iver-wharf/wharf-cmd/pkg/varsub"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml"
	"github.com/spf13/cobra"
	"gopkg.in/typ.v3/pkg/slices"
)

var varsSubstituteCmd = &cobra.Command{
	Use:     "substitute [path]",
	Aliases: []string{"sub"},
	Short:   "Replace values from lines piped in from STDIN",
	Long: `Performs variable substitution on each line that is piped in
from STDIN. Can be chained to make a new file with
variables substituted, like so:

	cat orig.txt | wharf vars substitute > new-file.txt`,
	Args: cobra.MaximumNArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"yml"}, cobra.ShellCompDirectiveFilterFileExt
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		currentDir, err := parseCurrentDir(slices.SafeGet(args, 0))
		if err != nil {
			return err
		}
		def, err := parseBuildDefinition(currentDir, wharfyml.Args{
			Env:    varsFlags.env,
			Inputs: parseInputArgs(varsFlags.inputs),
		})
		if err != nil {
			return err
		}
		copier := varsub.NewCopier(def.VarSource)
		copier.Copy(os.Stdout, os.Stdin)
		return nil
	},
}

func init() {
	varsCmd.AddCommand(varsSubstituteCmd)
}
