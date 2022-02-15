package cmd

import (
	"github.com/spf13/cobra"
)

var provisionerCmd = &cobra.Command{
	Use:   "provisioner",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
`,
}

func init() {
	rootCmd.AddCommand(provisionerCmd)
}
