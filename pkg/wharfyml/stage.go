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
	Name         string
	Environments []string
	Steps        []Step
}

func visitStageNode(name string, node ast.Node) (stage Stage, errSlice Errors) {
	stage.Name = name
	nodes, err := parseMappingValueNodes(node)
	if err != nil {
		errSlice.add(err)
		return
	}
	if len(nodes) == 0 {
		errSlice.add(newPositionedErrorNode(ErrStageEmpty, node))
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
			envs, errs := visitEnvironmentStringsNode(stepNode.Value)
			stage.Environments = envs
			errSlice.add(wrapPathErrorSlice(propEnvironments, errs)...)
		default:
			step, errs := visitStepNode(key, stepNode.Value)
			stage.Steps = append(stage.Steps, step)
			errSlice.add(wrapPathErrorSlice(key, errs)...)
		}
	}
	return
}
