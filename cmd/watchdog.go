package cmd

import (
	"github.com/iver-wharf/wharf-cmd/pkg/watchdog"
	"github.com/spf13/cobra"
)

var watchdogCmd = &cobra.Command{
	Use:   "watchdog",
	Short: "Watches for stray builds or workers, and kills them",
	Long: `The watchdog process will periodically query the wharf-api and
wharf-cmd-provisioner, perform a diff, and kill any builds or workers
that are missing from one another. Effectively killing all in the
symmetrical difference result.`,
	Run: func(cmd *cobra.Command, args []string) {
		watchdog.Watch()
	},
}

func init() {
	rootCmd.AddCommand(watchdogCmd)
}
