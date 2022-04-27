package main

import (
	"io"
	"os"

	"github.com/iver-wharf/wharf-cmd/pkg/varsub"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml"
	"github.com/spf13/cobra"
	"gopkg.in/typ.v3/pkg/slices"
)

var varsSubstituteCmd = &cobra.Command{
	Use:     "substitute [path]",
	Aliases: []string{"sub"},
	Short:   "Replace values from lines piped in from file or STDIN",
	Long: `Performs variable substitution on each line that is piped in
from STDIN, or from a file as provided by the second argument, and
writes the variable substituted values to STDOUT.

Can be chained to make a new file with variables substituted, like so:

	wharf vars substitute . orig-file.txt > new-file.txt`,
	Args: cobra.MaximumNArgs(2),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return []string{"yml"}, cobra.ShellCompDirectiveFilterFileExt
		}
		return nil, cobra.ShellCompDirectiveDefault
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
		var reader io.ReadCloser = os.Stdin
		if len(args) == 2 {
			file, err := os.Open(args[1])
			if err != nil {
				return err
			}
			reader = file
		}
		defer os.Stdout.Close()
		defer reader.Close()
		copier := varsub.NewCopier(def.VarSource)
		copier.Copy(os.Stdout, reader)
		return nil
	},
}

func init() {
	varsCmd.AddCommand(varsSubstituteCmd)
}
