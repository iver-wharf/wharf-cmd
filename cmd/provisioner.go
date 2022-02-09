package cmd

import (
	"context"

	"github.com/iver-wharf/wharf-cmd/pkg/provisioner"
	"github.com/spf13/cobra"
)

var provisionerCmd = &cobra.Command{
	Use:   "provisioner",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
`,
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := provisioner.NewK8sProvisioner("default", Kubeconfig)
		if err != nil {
			return err
		}
		return p.Serve(context.Background())
	},
}

func init() {
	provisionerCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(provisionerCmd)
}
