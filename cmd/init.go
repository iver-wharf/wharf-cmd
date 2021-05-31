package cmd

import (
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var initPath string
var context string

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		log.WithFields(log.Fields{"path": initPath, "context": context}).Debugln("init called")

		err := filepath.Walk(initPath, func(path string, f os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if f.IsDir() == false && f.Name() == "Dockerfile" {
				log.WithField("path", path).Debugln("Found dockerfile")
			}
			return nil
		})

		if err != nil {
			log.Errorln(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringVarP(&initPath, "path", "p", ".", "Path to repository")
	initCmd.Flags().StringVarP(&context, "context", "c", "", "Set a fixed context for Dockerfiles")
}
