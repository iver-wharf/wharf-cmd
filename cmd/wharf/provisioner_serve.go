package main

import (
	"github.com/iver-wharf/wharf-cmd/pkg/provisionerapi"
	"github.com/spf13/cobra"
)

var provisionerServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Starts serving HTTP REST API",
	Long: `Starts serving a HTTP REST API that the wharf-api uses to
provision new builds inside Kubernetes. The endpoints available are
equivalent to the "wharf provisioner" subcommands.

You can see an offline Swagger documentation of the API by visiting
the following URL path on a running wharf provisioner server:

	/api/swagger/index.html
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return provisionerapi.Serve(rootConfig.ProvisionerAPI, provisionerFlags.restConfig)
	},
}

func init() {
	provisionerCmd.AddCommand(provisionerServeCmd)
}
