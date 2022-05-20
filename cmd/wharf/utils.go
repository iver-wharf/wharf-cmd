package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/iver-wharf/wharf-cmd/internal/errutil"
	"github.com/iver-wharf/wharf-cmd/internal/flagtypes"
	"github.com/iver-wharf/wharf-cmd/internal/gitutil"
	"github.com/iver-wharf/wharf-cmd/internal/lastbuild"
	"github.com/iver-wharf/wharf-cmd/internal/util"
	"github.com/iver-wharf/wharf-cmd/pkg/steps"
	"github.com/iver-wharf/wharf-cmd/pkg/varsub"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml/visit"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/typ.v4/slices"
)

func parseCurrentDir(dirArg string) (string, error) {
	if dirArg == "" {
		return os.Getwd()
	}
	abs, err := filepath.Abs(dirArg)
	if err != nil {
		return "", err
	}
	stat, err := os.Stat(abs)
	if err != nil {
		return "", err
	}
	if !stat.IsDir() {
		dir, file := filepath.Split(abs)
		if file == ".wharf-ci.yml" {
			return filepath.Clean(dir), nil
		}
		return "", fmt.Errorf("path is neither a dir nor a .wharf-ci.yml file: %s", abs)
	}
	_, err = os.Stat(filepath.Join(abs, ".wharf-ci.yml"))
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("missing .wharf-ci.yml file in dir: %s", abs)
		}
		return "", err
	}
	return abs, nil
}

func parseVarSourcesExceptGit(currentDir string, additionalSource varsub.Source) (varsub.SourceSlice, error) {
	var varSources varsub.SourceSlice

	if additionalSource != nil {
		varSources = append(varSources, additionalSource)
	}

	varSources = append(varSources, varsub.NewOSEnvSource("WHARF_VAR_"))

	varFileSource, errs := wharfyml.ParseVarFiles(currentDir)
	if len(errs) > 0 {
		logParseErrors(errs, currentDir)
		return nil, errors.New("failed to parse variable files")
	}
	varSources = append(varSources, varFileSource)

	return varSources, nil
}

func parseVarSources(currentDir string, additionalSource varsub.Source) (varsub.Source, error) {
	varSources, err := parseVarSourcesExceptGit(currentDir, additionalSource)
	if err != nil {
		return nil, err
	}

	gitStats, err := gitutil.StatsFromExec(currentDir)
	if err != nil {
		log.Warn().WithError(err).
			Message("Failed to get REPO_ and GIT_ variables from Git. Skipping those.")
	} else {
		log.Debug().Message("Read REPO_ and GIT_ variables from Git:\n" +
			gitStats.String())
		varSources = append(varSources, gitStats)
	}

	return varSources, nil
}

func parseBuildDefinition(currentDir string, ymlArgs wharfyml.Args) (wharfyml.Definition, error) {
	ymlArgs.StepTypeFactory = steps.NewFactory(&rootConfig)

	varSource, err := parseVarSources(currentDir, ymlArgs.VarSource)
	if err != nil {
		return wharfyml.Definition{}, err
	}

	ymlArgs.VarSource = varSource

	ymlPath := filepath.Join(currentDir, ".wharf-ci.yml")
	log.Debug().WithString("path", ymlPath).Message("Parsing .wharf-ci.yml file.")
	def, errs := wharfyml.ParseFile(ymlPath, ymlArgs)
	if len(errs) > 0 {
		logParseErrors(errs, currentDir)
		return wharfyml.Definition{}, errors.New("failed to parse .wharf-ci.yml")
	}
	log.Debug().Message("Successfully parsed .wharf-ci.yml")
	return def, nil
}

func parseInputArgs(inputs flagtypes.KeyValueArray) map[string]any {
	m := make(map[string]any, len(inputs.Pairs))
	for _, kv := range inputs.Pairs {
		m[kv.Key] = kv.Value
	}
	return m
}

func logParseErrors(errs errutil.Slice, currentDir string) {
	log.Warn().WithInt("errors", len(errs)).Message("Cannot run build due to parsing errors.")
	log.Warn().Message("")
	for _, err := range errs {
		scopePrefix := errutil.AsScope(err)
		if scopePrefix != "" {
			scopePrefix += ": "
		}
		var posErr errutil.Pos
		if errors.As(err, &posErr) {
			log.Warn().Messagef("%4d:%-4d %s%s",
				posErr.Line, posErr.Column, scopePrefix, err.Error())
		} else {
			log.Warn().Messagef("   -:-    %s%s", scopePrefix, err.Error())
		}
	}
	log.Warn().Message("")

	containsMissingBuiltin := false
	for _, err := range errs {
		if errors.Is(err, visit.ErrMissingBuiltinVar) {
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
			if file.IsRel {
				continue
			}
			sb.WriteString("\n  ")
			sb.WriteString(file.PrettyPath(currentDir))
		}
		sb.WriteString(`

Wharf also looks for:
  - All ".wharf-vars.yml" in this directory or any parent directory.
  - Local Git repository and extracts GIT_ and REPO_ variables from it.
  - Environment variables, with prefix WHARF_VAR_ removed from them.

Sample file content:
  # .wharf-vars.yml
  vars:
    REG_URL: http://harbor.example.com
`)
		log.Info().Message(sb.String())
	}
}

func addWharfYmlStageFlag(cmd *cobra.Command, flags *pflag.FlagSet, value *string) {
	flags.StringVarP(value, "stage", "s", "", "Stage to run (will run all stages if unset)")
	cmd.RegisterFlagCompletionFunc("stage", completeWharfYmlStage)
}

func addWharfYmlEnvFlag(cmd *cobra.Command, flags *pflag.FlagSet, value *string) {
	flags.StringVarP(value, "environment", "e", "", "Environment selection")
	cmd.RegisterFlagCompletionFunc("environment", completeWharfYmlEnv)
}

func addWharfYmlInputsFlag(cmd *cobra.Command, flags *pflag.FlagSet, value *flagtypes.KeyValueArray) {
	flags.VarP(value, "input", "i", "Inputs (--input key=value), can be set multiple times")
	cmd.RegisterFlagCompletionFunc("input", completeWharfYmlInputs)
}

func completeWharfYmlStage(cmd *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
	def, err := parseWharfYmlForCompletions(cmd, args)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var stages []string
	for _, stage := range def.Stages {
		stages = append(stages, stage.Name)
	}
	return stages, cobra.ShellCompDirectiveNoFileComp
}

func completeWharfYmlEnv(cmd *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
	def, err := parseWharfYmlForCompletions(cmd, args)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var envs []string
	for env := range def.Envs {
		envs = append(envs, env)
	}
	return envs, cobra.ShellCompDirectiveNoFileComp
}

func completeWharfYmlInputs(cmd *cobra.Command, args []string, completed string) ([]string, cobra.ShellCompDirective) {
	def, err := parseWharfYmlForCompletions(cmd, args)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var inputs []string
	for _, input := range def.Inputs {
		inputs = append(inputs, input.InputVarName())
	}
	return flagtypes.CompleteKeyValue(inputs, completed)
}

func parseWharfYmlForCompletions(cmd *cobra.Command, args []string) (wharfyml.Definition, error) {
	currentDir, err := parseCurrentDir(slices.SafeGet(args, 0))
	if err != nil {
		return wharfyml.Definition{}, err
	}

	var ymlArgs wharfyml.Args
	envFlag := cmd.Flag("environment")
	if envFlag.Changed {
		ymlArgs.Env = envFlag.Value.String()
	} else {
		ymlArgs.SkipStageFiltering = true
	}

	ymlPath := filepath.Join(currentDir, ".wharf-ci.yml")

	// Intentionally ignore any parse errors, as syntax errors or missing fields
	// for a step type are irrelevant for the completions.
	def, _ := wharfyml.ParseFile(ymlPath, ymlArgs)
	return def, nil
}

type commonVarSubFlags struct {
	buildID   uint
	projectID uint
}

func (flags commonVarSubFlags) varSource() varsub.Source {
	m := make(varsub.SourceMap)
	if flags.buildID != 0 {
		sourceName := "flag --build-id"
		if path, err := lastbuild.Path(); err == nil {
			sourceName = fmt.Sprintf(
				"%s, or next ID from %s", sourceName, util.ShorthandHome(path))
		}
		m["BUILD_REF"] = varsub.Val{
			Value:  flags.buildID,
			Source: sourceName,
		}
	}
	if flags.projectID != 0 {
		m["PROJECT_ID"] = varsub.Val{
			Value:  flags.projectID,
			Source: "flag --project-id",
		}
	}
	return m
}

func addCommonVarSubFlags(flags *pflag.FlagSet, varFlags *commonVarSubFlags) {
	flags.UintVar(&varFlags.projectID, "project-id", 0, "Overrides PROJECT_ID variable")

	buildIDHelp := "Overrides BUILD_REF variable"
	if path, err := lastbuild.Path(); err == nil {
		if nextGuess, err := lastbuild.GuessNext(); err == nil {
			buildIDHelp = fmt.Sprintf("%s (default %d, via %q)", buildIDHelp, nextGuess, path)
		}
	}
	flags.UintVar(&varFlags.buildID, "build-id", 0, buildIDHelp)
}
