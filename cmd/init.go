package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initPath string
var context string

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug().
			WithString("path", initPath).
			WithString("context", context).
			Message("init called")

		err := filepath.Walk(initPath, func(path string, f os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if f.IsDir() == false && f.Name() == "Dockerfile" {
				log.Debug().WithString("path", path).Message("Found Dockerfile.")
			}
			return nil
		})

		if err != nil {
			log.Error().WithError(err).Message("")
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringVarP(&initPath, "path", "p", ".", "Path to repository")
	initCmd.Flags().StringVarP(&context, "context", "c", "", "Set a fixed context for Dockerfiles")
}
