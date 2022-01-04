package cmd

import (
	"strings"

	"github.com/spf13/cobra"
	pkgns "github.com/iver-wharf/wharf-cmd/pkg/namespace"
)

var namespaces []string

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "A brief description of your command",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug().WithString("namespaces", strings.Join(namespaces, ",")).Message("setup called")

		pkgns.Namespaces{Kubeconfig: Kubeconfig}.SetupNamespaces(namespaces)
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)

	setupCmd.Flags().StringArrayVarP(&namespaces, "namespaces", "n", []string{"default"}, "Namespaces to add deploy user to (creates namespaces if not exists)")
}
