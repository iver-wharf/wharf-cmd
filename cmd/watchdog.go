package cmd

import (
	"github.com/spf13/cobra"
)

var watchdogCmd = &cobra.Command{
	Use:   "watchdog",
	Short: "Watchdog monitors stray builds or workers",
	Long: `The watchdog tool is used to monitor both wharf-api and
wharf-cmd-provisioner to see if a build or worker exists in one
but not the other, and then kills those stray builds or workers.`,
}

func init() {
	rootCmd.AddCommand(watchdogCmd)
}
