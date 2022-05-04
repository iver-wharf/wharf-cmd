package main

import (
	"github.com/iver-wharf/wharf-cmd/pkg/watchdog"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Kills stray builds and workers on an interval",
	Long: `Serve will periodically query the wharf-api and
wharf-cmd-provisioner, perform a diff, and kill any builds or workers
that are missing from one another. Effectively killing all in the
symmetrical difference result.`,
	Run: func(cmd *cobra.Command, args []string) {
		watchdog.Watch(rootConfig.Watchdog)
	},
}

func init() {
	watchdogCmd.AddCommand(serveCmd)
}
