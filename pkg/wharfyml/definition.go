package wharfyml

import (
	"errors"
	"fmt"

	"github.com/iver-wharf/wharf-cmd/internal/errutil"
	"github.com/iver-wharf/wharf-cmd/pkg/varsub"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml/visit"
	"gopkg.in/yaml.v3"
)

// Errors specific to parsing definitions.
var (
	ErrUseOfUndefinedEnv = errors.New("use of undefined environment")
)

// Definition is the .wharf-ci.yml build definition structure.
type Definition struct {
	Inputs    Inputs
	Envs      map[string]Env
	Env       *Env
	Stages    []Stage
	VarSource varsub.Source
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

func visitDefNode(node *yaml.Node, args Args) (def Definition, errSlice errutil.Slice) {
	nodes, errs := visit.MapSlice(node)
	errSlice.Add(errs...)
	envSourceNode := node

	for _, n := range nodes {
		switch n.Key.Value {
		case propEnvironments:
			var errs errutil.Slice
			def.Envs, errs = visitDocEnvironmentsNode(n.Value)
			errSlice.Add(errutil.ScopeSlice(errs, propEnvironments)...)
			envSourceNode = n.Value
		case propInputs:
			var errs errutil.Slice
			def.Inputs, errs = visitInputsNode(n.Value)
			errSlice.Add(errutil.ScopeSlice(errs, propInputs)...)
		}
	}

	var sources varsub.SourceSlice

	inputsSource, errs := visitInputsArgs(def.Inputs, args.Inputs)
	sources = append(sources, inputsSource)
	errSlice.Add(errs...)

	sources = append(sources, def.Inputs.DefaultsVarSource())

	// Add environment varsub.Source first, as it should have priority
	targetEnv, err := getTargetEnv(def.Envs, args.Env)
	if err != nil {
		err = errutil.NewPosFromNode(err, envSourceNode)
		err = errutil.Scope(err, propEnvironments)
		errSlice.Add(err) // Non fatal error
	} else if targetEnv != nil {
		def.Env = targetEnv
		sources = append(sources, targetEnv.VarSource())
	}

	if args.VarSource != nil {
		sources = append(sources, args.VarSource)
	}

	stages, errs := visitDefStageNodes(nodes, args, sources)
	def.Stages = stages
	errSlice.Add(errs...)
	errSlice.Add(validateDefEnvironmentUsage(def)...)
	if !args.SkipStageFiltering {
		// filtering intentionally performed after validation
		def.Stages = filterStagesOnEnv(def.Stages, args.Env)
	}
	def.VarSource = sources
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

func visitDefStageNodes(nodes []visit.MapItem, args Args, source varsub.Source) (stages []Stage, errSlice errutil.Slice) {
	for _, n := range nodes {
		switch n.Key.Value {
		case propEnvironments, propInputs:
			// Do nothing, they've already been visited.
			continue
		}
		stageNode, err := visit.VarSubNodeRec(n.Value, source)
		if err != nil {
			errSlice.Add(err)
			continue
		}
		stage, errs := visitStageNode(n.Key, stageNode, args, source)
		stages = append(stages, stage)
		errSlice.Add(errutil.ScopeSlice(errs, n.Key.Value)...)
	}
	return
}

func validateDefEnvironmentUsage(def Definition) errutil.Slice {
	var errSlice errutil.Slice
	for _, stage := range def.Stages {
		for _, env := range stage.Envs {
			if _, ok := def.Envs[env.Name]; !ok {
				err := fmt.Errorf("%w: %q", ErrUseOfUndefinedEnv, env.Name)
				err = errutil.NewPos(err, env.Source.Line, env.Source.Column)
				err = errutil.Scope(err, stage.Name, propEnvironments)
				errSlice.Add(err)
			}
		}
	}
	return errSlice
}

func filterStagesOnEnv(stages []Stage, envFilter string) []Stage {
	var filtered []Stage
	for _, stage := range stages {
		if stageShouldBeIncluded(stage.Envs, envFilter) {
			filtered = append(filtered, stage)
		}
	}
	return filtered
}

func stageShouldBeIncluded(envRefs []EnvRef, envFilter string) bool {
	if len(envRefs) == 0 {
		return true
	}
	if len(envRefs) > 0 && envFilter == "" {
		return false
	}
	for _, ref := range envRefs {
		if ref.Name == envFilter {
			return true
		}
	}
	return false
}
