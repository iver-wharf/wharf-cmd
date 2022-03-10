package wharfyml

import (
	"errors"
	"fmt"

	"github.com/iver-wharf/wharf-cmd/pkg/varsub"
	"gopkg.in/yaml.v3"
)

// Errors specific to parsing definitions.
var (
	ErrUseOfUndefinedEnv = errors.New("use of undefined environment")
)

// Definition is the .wharf-ci.yml build definition structure.
type Definition struct {
	Inputs map[string]Input
	Envs   map[string]Env
	Env    *Env
	Stages []Stage
}

// ListAllSteps aggregates steps from all stages into a single slice.
//
// Makes Definition comply to the StepLister interface.
func (d *Definition) ListAllSteps() []Step {
	var steps []Step
	for _, stage := range d.Stages {
		steps = append(steps, stage.Steps...)
	}
	return steps
}

func visitDefNode(node *yaml.Node, args Args) (def Definition, errSlice Errors) {
	nodes, errs := visitMapSlice(node)
	errSlice.add(errs...)
	envSourceNode := node
	for _, n := range nodes {
		switch n.key.value {
		case propEnvironments:
			var errs Errors
			def.Envs, errs = visitDocEnvironmentsNode(n.value)
			errSlice.add(wrapPathErrorSlice(errs, propEnvironments)...)
			envSourceNode = n.value
		case propInputs:
			var errs Errors
			def.Inputs, errs = visitInputsNode(n.value)
			errSlice.add(wrapPathErrorSlice(errs, propInputs)...)
		}
	}

	var sources varsub.SourceSlice
	if args.VarSource != nil {
		sources = append(sources, args.VarSource)
	}

	targetEnv, err := getTargetEnv(def.Envs, args.Env)
	if err != nil {
		err = wrapPosErrorNode(err, envSourceNode)
		err = wrapPathError(err, propEnvironments)
		errSlice.add(err) // Non fatal error
	} else if targetEnv != nil {
		def.Env = targetEnv
		sources = append(sources, varsub.SourceMap(targetEnv.Vars))
	}

	stages, errs := visitDefStageNodes(nodes, sources)
	def.Stages = stages
	errSlice.add(errs...)
	errSlice.add(validateDefEnvironmentUsage(def)...)
	// filtering intentionally performed after validation
	def.Stages = filterStagesOnEnv(def.Stages, args.Env)
	return
}

func getTargetEnv(envs map[string]Env, envName string) (*Env, error) {
	if envName == "" {
		return nil, nil
	}
	env, ok := envs[envName]
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrUseOfUndefinedEnv, envName)
	}
	return &env, nil
}

func visitDefStageNodes(nodes []mapItem, source varsub.Source) (stages []Stage, errSlice Errors) {
	for _, n := range nodes {
		switch n.key.value {
		case propEnvironments, propInputs:
			// Do nothing, they've already been visited.
			continue
		}
		stageNode, err := varSubNodeRec(n.value, source)
		if err != nil {
			errSlice.add(err)
			continue
		}
		stage, errs := visitStageNode(n.key, stageNode)
		stages = append(stages, stage)
		errSlice.add(wrapPathErrorSlice(errs, n.key.value)...)
	}
	return
}

func validateDefEnvironmentUsage(def Definition) Errors {
	var errSlice Errors
	for _, stage := range def.Stages {
		for _, env := range stage.Envs {
			if _, ok := def.Envs[env.Name]; !ok {
				err := fmt.Errorf("%w: %q", ErrUseOfUndefinedEnv, env.Name)
				err = wrapPosError(err, env.Source)
				err = wrapPathError(err, stage.Name, propEnvironments)
				errSlice.add(err)
			}
		}
	}
	return errSlice
}

func filterStagesOnEnv(stages []Stage, envFilter string) []Stage {
	var filtered []Stage
	for _, stage := range stages {
		if containsEnvFilter(stage.Envs, envFilter) {
			filtered = append(filtered, stage)
		}
	}
	return filtered
}

func containsEnvFilter(envRefs []EnvRef, envFilter string) bool {
	if envFilter == "" && len(envRefs) == 0 {
		return true
	}
	for _, ref := range envRefs {
		if ref.Name == envFilter {
			return true
		}
	}
	return false
}
