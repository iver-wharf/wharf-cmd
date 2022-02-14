package wharfyml

import (
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"
)

var (
	ErrUseOfUndefinedEnv = errors.New("use of undefined environment")
)

// Definition is the .wharf-ci.yml build definition structure.
type Definition struct {
	Inputs map[string]Input
	Envs   map[string]Env
	Stages []Stage
}

func visitDefNode(node *yaml.Node) (def Definition, errSlice Errors) {
	nodes, errs := visitMapSlice(node)
	errSlice.add(errs...)
	for _, n := range nodes {
		switch n.key.value {
		case propEnvironments:
			var errs Errors
			def.Envs, errs = visitDocEnvironmentsNode(n.value)
			errSlice.add(wrapPathErrorSlice(propEnvironments, errs)...)
		case propInputs:
			var errs Errors
			def.Inputs, errs = visitInputsNode(n.value)
			errSlice.add(wrapPathErrorSlice(propInputs, errs)...)
		default:
			stage, errs := visitStageNode(n.key, n.value)
			def.Stages = append(def.Stages, stage)
			errSlice.add(wrapPathErrorSlice(n.key.value, errs)...)
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
				errSlice.add(wrapPosError(err, env.Source))
			}
		}
	}
	return errSlice
}
