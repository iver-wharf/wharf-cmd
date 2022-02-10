package wharfyml

import (
	"errors"

	"github.com/goccy/go-yaml/ast"
)

// Errors related to parsing stages.
var (
	ErrStageEmpty = errors.New("stage is missing steps")
)

// Stage holds the name, environment filter, and list of steps for this Wharf
// build stage.
type Stage struct {
	Pos     Pos
	Name    string
	Envs    []string
	EnvsPos Pos
	Steps   []Step
}

func visitStageNode(name string, node ast.Node) (stage Stage, errSlice Errors) {
	stage.Pos = newPosNode(node)
	stage.Name = name
	nodes, err := parseMappingValueNodes(node)
	if err != nil {
		errSlice.add(err)
		return
	}
	if len(nodes) == 0 {
		errSlice.add(wrapPosErrorNode(ErrStageEmpty, node))
		return
	}
	for _, stepNode := range nodes {
		key, err := parseMapKeyNonEmpty(stepNode.Key)
		if err != nil {
			errSlice.add(err)
			// non-fatal error
		}
		switch key {
		case propEnvironments:
			stage.EnvsPos = newPosNode(stepNode.Value)
			envs, errs := visitStageEnvironmentsNode(stepNode.Value)
			stage.Envs = envs
			errSlice.add(wrapPathErrorSlice(propEnvironments, errs)...)
		default:
			step, errs := visitStepNode(key, stepNode.Value)
			stage.Steps = append(stage.Steps, step)
			errSlice.add(wrapPathErrorSlice(key, errs)...)
		}
	}
	return
}
