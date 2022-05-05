package main

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/fatih/color"
	"github.com/iver-wharf/wharf-cmd/pkg/varsub"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml"
	"github.com/spf13/cobra"
	"gopkg.in/typ.v4/slices"
)

var (
	colorVarSourceName     = color.New(color.FgYellow)
	colorVarKey            = color.New(color.FgHiMagenta)
	colorVarValue          = color.New()
	colorVarOverridden     = color.New(color.FgHiBlack, color.CrossedOut)
	colorVarOverriddenNote = color.New(color.FgHiBlack, color.Italic)
)

var varsListFlags = struct {
	showAll bool
}{}

var varsListCmd = &cobra.Command{
	Use:     "list [path]",
	Aliases: []string{"ls"},
	Short:   "Print all variables that would be used for a .wharf-ci.yml file",
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
			Env:    varsFlags.env,
			Inputs: parseInputArgs(varsFlags.inputs),
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

		type variable struct {
			varsub.Var
			isUsed bool
		}

		for _, g := range groups {
			slices.SortFunc(g.Values, func(a, b varsub.Var) bool {
				return a.Key < b.Key
			})
			var vars []variable
			for _, value := range g.Values {
				v, _ := def.VarSource.Lookup(value.Key)
				vars = append(vars, variable{
					Var:    v,
					isUsed: v.Source == value.Source,
				})
			}

			var longestKeyLength int
			var notUsedCount int
			for _, v := range vars {
				if !varsListFlags.showAll && !v.isUsed {
					notUsedCount++
					continue
				}
				if len(v.Key) > longestKeyLength {
					longestKeyLength = len(v.Key)
				}
			}

			if g.Key == "" {
				colorVarSourceName.Fprint(&sb, "\n(undefined source):\n")
			} else {
				colorVarSourceName.Fprintf(&sb, "\n%s:\n", g.Key)
			}

			longestSpaces := strings.Repeat(" ", longestKeyLength+2)
			for _, v := range vars {
				if !varsListFlags.showAll && !v.isUsed {
					continue
				}
				spacesCount := longestKeyLength - utf8.RuneCountInString(v.Key) + 2
				spaces := longestSpaces[:spacesCount]
				sb.WriteString("  ")
				if v.isUsed {
					fmt.Fprintf(&sb, "%s%s%s\n",
						colorVarKey.Sprint(v.Key),
						spaces,
						colorVarValue.Sprint(v.Value))
				} else {
					colorVarOverridden.Fprintf(&sb, "%s%s%s\n", v.Key, spaces, v.Value)
				}
			}

			if !varsListFlags.showAll && notUsedCount > 0 {
				sb.WriteString("  ")
				colorVarOverriddenNote.Fprintf(&sb, "(hiding %d overridden variables)\n", notUsedCount)
			}
		}

		log.Info().Message(sb.String())

		return nil
	},
}

func init() {
	varsCmd.AddCommand(varsListCmd)

	varsListCmd.PersistentFlags().BoolVarP(&varsListFlags.showAll, "all", "a", false, "Show overridden variables")
}
