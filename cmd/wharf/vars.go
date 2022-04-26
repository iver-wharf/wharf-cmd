package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/iver-wharf/wharf-cmd/pkg/varsub"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml"
	"github.com/spf13/cobra"
	"gopkg.in/typ.v3/pkg/slices"
)

var varsFlags = struct {
	env string
}{}

var varsCmd = &cobra.Command{
	Use:   "vars [path]",
	Short: "Print all variables that would be used for a .wharf-ci.yml file",
	Long: `Parses a .wharf-ci.yml and all .wharf-vars.yml files as if it was
running "wharf run", but prints out all the variables that would be used
instead of performing the build.

The variables sources are printed in the order of priority, where the latter
sources override the former sources, if a variable would have the same name
in multiple sources.`,
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
			Env: varsFlags.env,
		})
		if err != nil {
			return err
		}

		vars := def.VarSource.ListVars()

		if len(vars) == 0 {
			if len(def.Envs) > 0 {
				var env wharfyml.Env
				for _, e := range def.Envs {
					env = e
					break
				}
				log.Info().Messagef(`No variables found.

Try specifying the --environment flag to include
environment variables, such as:

  %s vars --environment %q`, os.Args[0], env.Name)
				return nil
			}

			log.Info().Messagef("No variables found.")
			return nil
		}

		groups := slices.GroupBy(vars, func(v varsub.Var) string {
			return v.Source
		})
		slices.Reverse(groups)

		var sb strings.Builder
		fmt.Fprintf(&sb, "Found %d variables from %d different sources:\n", len(vars), len(groups))

		for _, g := range groups {
			slices.SortFunc(g.Values, func(a, b varsub.Var) bool {
				return a.Key < b.Key
			})

			var longestKeyLength int
			for _, v := range g.Values {
				if len(v.Key) > longestKeyLength {
					longestKeyLength = len(v.Key)
				}
			}

			if g.Key == "" {
				sb.WriteString("\n(undefined source):\n")
			} else {
				fmt.Fprintf(&sb, "\n%s:\n", g.Key)
			}

			format := fmt.Sprintf("  %%%ds %%v\n", -longestKeyLength-1)
			for _, v := range g.Values {
				fmt.Fprintf(&sb, format, v.Key, v.Value)
			}
		}

		log.Info().Message(sb.String())

		return nil
	},
}

func init() {
	rootCmd.AddCommand(varsCmd)

	varsCmd.Flags().StringVarP(&varsFlags.env, "environment", "e", "", "Environment selection")
}
