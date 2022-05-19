package main

import (
	"os"

	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml/visit"
	"github.com/spf13/cobra"
	"gopkg.in/typ.v4/slices"
	"gopkg.in/yaml.v3"
)

var varsYMLFlags = struct {
	stage string
}{}

var varsYMLCmd = &cobra.Command{
	Use:     "yml [path]",
	Aliases: []string{"yaml"},
	Short:   "Print the parsed .wharf-ci.yml file",
	Long: `Parses a .wharf-ci.yml file and prints the parsed definition
how it would be used if all values were inlined.`,
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

		writer := os.Stdout
		for _, stage := range def.Stages {
			enc := yaml.NewEncoder(writer)
			enc.SetIndent(2)
			enc.Encode(yaml.Node{
				Kind: yaml.MappingNode,
				Tag:  visit.ShortTagMap,
				Content: []*yaml.Node{
					stage.Node.Key.Node,
					stage.Node.Value,
				},
			})
			writer.WriteString("\n")
		}

		writer.Close()

		return nil
	},
}

func init() {
	varsCmd.AddCommand(varsYMLCmd)

	addWharfYmlStageFlag(varsYMLCmd, varsYMLCmd.Flags(), &varsYMLFlags.stage)
}
