package wharfyml

import (
	"errors"
	"fmt"

	"github.com/goccy/go-yaml/ast"
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

func visitDefNodes(nodes []*ast.MappingValueNode) (def Definition, errSlice Errors) {
	for _, n := range nodes {
		key, err := parseMapKeyNonEmpty(n.Key)
		if err != nil {
			errSlice.add(fmt.Errorf("%q: %w", n.Key, err))
			// non-fatal error
		}
		switch key {
		case propEnvironments:
			var errs Errors
			def.Envs, errs = visitDocEnvironmentsNode(n.Value)
			errSlice.add(wrapPathErrorSlice(propEnvironments, errs)...)
		case propInputs:
			var errs Errors
			def.Inputs, errs = visitInputsNode(n.Value)
			errSlice.add(wrapPathErrorSlice(propInputs, errs)...)
		default:
			stage, errs := visitStageNode(key, n.Value)
			def.Stages = append(def.Stages, stage)
			errSlice.add(wrapPathErrorSlice(key, errs)...)
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
