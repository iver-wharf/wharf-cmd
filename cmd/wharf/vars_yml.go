package main

import (
	"os"

	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml/visit"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var varsYMLFlags = struct {
	stage   string
	showAll bool
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
		def, err := varsCmdParseBuildDef(args)
		if err != nil {
			return err
		}

		writer := os.Stdout
		for _, stage := range def.Stages {
			if !varsYMLFlags.showAll &&
				varsYMLFlags.stage != "" &&
				varsYMLFlags.stage != stage.Name {
				continue
			}
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
	varsYMLCmd.Flags().BoolVarP(&varsYMLFlags.showAll, "all", "a", false, "Show all stages, skipping stage and environment filtering")
}
