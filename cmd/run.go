package cmd

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/iver-wharf/wharf-cmd/pkg/gitstat"
	"github.com/iver-wharf/wharf-cmd/pkg/resultstore"
	"github.com/iver-wharf/wharf-cmd/pkg/varsub"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml"
	"github.com/iver-wharf/wharf-cmd/pkg/worker"
	"github.com/iver-wharf/wharf-cmd/pkg/worker/workermodel"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

var runFlags = struct {
	path         string
	stage        string
	env          string
	k8sOverrides clientcmd.ConfigOverrides
}{}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Runs a new build from a .wharf-ci.yml file",
	Long: `Runs a new build in a Kubernetes cluster using pods
based on a .wharf-ci.yml file.

If no stage is specified via --stage then wharf-cmd will run all stages
in sequence, based on their order of declaration in the .wharf-ci.yml file.

All steps in each stage will be run in parallel for each stage.

Read more about the .wharf-ci.yml file here:
https://iver-wharf.github.io/#/usage-wharfyml/
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		kubeconfig, ns, err := loadKubeconfig(runFlags.k8sOverrides)
		if err != nil {
			return err
		}
		ymlAbsPath, err := filepath.Abs(runFlags.path)
		if err != nil {
			return fmt.Errorf("get absolute path of .wharf-ci.yml file: %w", err)
		}
		currentDir := filepath.Dir(ymlAbsPath)

		var varSources varsub.SourceSlice

		varFileSource, errs := wharfyml.ParseVarFiles(currentDir)
		if len(errs) > 0 {
			logParseErrors(errs, currentDir)
			return errors.New("failed to parse variable files")
		}
		if varFileSource != nil {
			varSources = append(varSources, varFileSource)
		}

		gitStats, err := gitstat.FromExec(currentDir)
		if err != nil {
			log.Warn().WithError(err).
				Message("Failed to get REPO_ and GIT_ variables from Git. Skipping those.")
		} else {
			log.Debug().Message("Read REPO_ and GIT_ variables from Git:\n" +
				gitStats.String())
			varSources = append(varSources, gitStats)
		}

		varSources = append(varSources, varsub.EnvSource{})

		def, errs := wharfyml.ParseFile(ymlAbsPath, wharfyml.Args{
			Env:       runFlags.env,
			VarSource: varSources,
		})
		if len(errs) > 0 {
			logParseErrors(errs, currentDir)
			return errors.New("failed to parse .wharf-ci.yml")
		}
		log.Debug().Message("Successfully parsed .wharf-ci.yml")
		// TODO: Change to build ID-based path, e.g /tmp/iver-wharf/wharf-cmd/builds/123/...
		//
		// May require setting of owner and SUID on wharf-cmd binary to access /var/log.
		// e.g.:
		//  chown root $(which wharf-cmd) && chmod +4000 $(which wharf-cmd)
		store := resultstore.NewStore(resultstore.NewFS("/var/log/build_logs"))
		b, err := worker.NewK8s(context.Background(), def, ns, kubeconfig, store, worker.BuildOptions{
			StageFilter: runFlags.stage,
		})
		if err != nil {
			return err
		}
		log.Debug().Message("Successfully created builder.")
		log.Info().Message("Starting build.")
		res, err := b.Build(context.Background())
		if err != nil {
			return err
		}
		log.Info().
			WithDuration("dur", res.Duration.Truncate(time.Second)).
			WithStringer("status", res.Status).
			Message("Done with build.")
		if res.Status != workermodel.StatusSuccess {
			return errors.New("build failed")
		}
		return nil
	},
}

func logParseErrors(errs wharfyml.Errors, currentDir string) {
	log.Warn().WithInt("errors", len(errs)).Message("Cannot run build due to parsing errors.")
	for _, err := range errs {
		var posErr wharfyml.PosError
		if errors.As(err, &posErr) {
			log.Warn().Messagef("%4d:%-4d%s",
				posErr.Source.Line, posErr.Source.Column, err.Error())
		} else {
			log.Warn().Messagef("   -:-   %s", err.Error())
		}
	}

	containsMissingBuiltin := false
	for _, err := range errs {
		if errors.Is(err, wharfyml.ErrMissingBuiltinVar) {
			containsMissingBuiltin = true
			break
		}
	}
	if containsMissingBuiltin {
		varFiles := wharfyml.ListPossibleVarsFiles(currentDir)
		var sb strings.Builder
		sb.WriteString(`Tip: You can add built-in variables to Wharf in multiple ways.

Wharf look for values in the following files:`)
		for _, file := range varFiles {
			if file.Kind == wharfyml.VarFileKindParentDir {
				continue
			}
			sb.WriteString("\n  ")
			sb.WriteString(file.PrettyPath(currentDir))
		}
		sb.WriteString(`

Wharf also looks for:
  - All ".wharf-vars.yml" in this directory or any parent directory.
  - Local Git repository and extracts GIT_ and REPO_ variables from it.
  - Environment variables.

Sample file content:
  # .wharf-vars.yml
  vars:
    REG_URL: http://harbor.example.com
`)
		log.Info().Message(sb.String())
	}
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().StringVarP(&runFlags.path, "path", "p", ".wharf-ci.yml", "Path to .wharf-ci file")
	runCmd.Flags().StringVarP(&runFlags.stage, "stage", "s", "", "Stage to run (will run all stages if unset)")
	runCmd.Flags().StringVarP(&runFlags.env, "environment", "e", "", "Environment selection")
	addKubernetesFlags(runCmd.Flags(), &runFlags.k8sOverrides)
}
