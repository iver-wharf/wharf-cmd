package wharfyml

import (
	"fmt"

	"github.com/goccy/go-yaml/ast"
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
			def.Envs, errs = visitDefEnvironmentsNodes(n.Value)
			errSlice.add(errs...)
		case propInputs:
			var errs Errors
			def.Inputs, errs = visitInputsNode(n.Value)
			errSlice.add(wrapPathErrorSlice(propInputs, errs)...)
		default:
			stage, errs := visitDefStageNode(key, n.Value)
			def.Stages = append(def.Stages, stage)
			errSlice.add(errs...)
		}
	}
	return
}

func visitDefEnvironmentsNodes(node ast.Node) (map[string]Env, Errors) {
	envs, errs := visitDocEnvironmentsNode(node)
	errs = wrapPathErrorSlice(propEnvironments, errs)
	return envs, errs
}

func visitDefStageNode(key string, node ast.Node) (Stage, Errors) {
	stage, errs := visitStageNode(key, node)
	errs = wrapPathErrorSlice(key, errs)
	return stage, errs
}
