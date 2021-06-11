package cmd

import (
	"github.com/spf13/cobra"
	"github.com/iver-wharf/wharf-cmd/pkg/serve"
)

var listen string

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		server := serve.Server{
			Kubeconfig: Kubeconfig,
			Namespace:  Namespace,
		}

		server.Serve(listen)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().StringVarP(&listen, "listen", "l", ":8080", "Listen string (eg. :8080")
}
