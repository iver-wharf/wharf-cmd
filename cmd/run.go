package cmd

import (
	"context"
	"errors"
	"time"

	"github.com/iver-wharf/wharf-cmd/pkg/builder"
	"github.com/iver-wharf/wharf-cmd/pkg/core/containercreator"
	"github.com/iver-wharf/wharf-cmd/pkg/core/wharfyml"
	"github.com/spf13/cobra"
)

var flagRunPath string

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		stepRun, err := builder.NewK8sStepRunner("build", Kubeconfig)
		if err != nil {
			return err
		}
		b := builder.New(stepRun)
		def, err := wharfyml.Parse(flagRunPath, make(map[containercreator.BuiltinVar]string))
		if err != nil {
			return err
		}
		res, err := b.Build(context.Background(), def)
		if err != nil {
			return err
		}
		log.Info().
			WithDuration("dur", res.Duration.Truncate(time.Second)).
			WithStringer("status", res.Status).
			Message("Done with build.")
		if res.Status != builder.StatusSuccess {
			return errors.New("build failed")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().StringVarP(&flagRunPath, "path", "p", ".wharf-ci.yml", "Path to .wharf-ci file")
}
