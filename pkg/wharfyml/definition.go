package wharfyml

import (
	"errors"
	"fmt"

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
	Stages []Stage
}

// ListBuildSteps aggregates steps from all stages into a single slice.
//
// Makes Definition comply to the BuildStepLister interface.
func (d *Definition) ListBuildSteps() []Step {
	var steps []Step
	for _, stage := range d.Stages {
		steps = append(steps, stage.Steps...)
	}
	return steps
}

func visitDefNode(node *yaml.Node) (def Definition, errSlice Errors) {
	nodes, errs := visitMapSlice(node)
	errSlice.add(errs...)
	for _, n := range nodes {
		switch n.key.value {
		case propEnvironments:
			var errs Errors
			def.Envs, errs = visitDocEnvironmentsNode(n.value)
			errSlice.add(wrapPathErrorSlice(errs, propEnvironments)...)
		case propInputs:
			var errs Errors
			def.Inputs, errs = visitInputsNode(n.value)
			errSlice.add(wrapPathErrorSlice(errs, propInputs)...)
		default:
			stage, errs := visitStageNode(n.key, n.value)
			def.Stages = append(def.Stages, stage)
			errSlice.add(wrapPathErrorSlice(errs, n.key.value)...)
		}
	}
	errSlice.add(validateDefEnvironmentUsage(def)...)
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
