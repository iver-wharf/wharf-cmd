package cmd

import (
	"github.com/iver-wharf/wharf-cmd/pkg/provisionerapi"
	"github.com/spf13/cobra"
)

var provisionerServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return provisionerapi.Serve(Kubeconfig)
	},
}
