package wharfyml

import (
	"errors"

	"github.com/iver-wharf/wharf-cmd/pkg/varsub"
	"gopkg.in/yaml.v3"
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
	Envs    []EnvRef
	EnvsPos Pos
	Steps   []Step
}

func visitStageNode(nameNode strNode, node *yaml.Node, source varsub.Source) (stage Stage, errSlice Errors) {
	stage.Pos = newPosNode(node)
	stage.Name = nameNode.value
	nodes, errs := visitMapSlice(node)
	errSlice.add(errs...)
	if len(nodes) == 0 {
		errSlice.add(wrapPosErrorNode(ErrStageEmpty, node))
		return
	}
	for _, stepNode := range nodes {
		switch stepNode.key.value {
		case propEnvironments:
			stage.EnvsPos = newPosNode(stepNode.value)
			envs, errs := visitStageEnvironmentsNode(stepNode.value)
			stage.Envs = envs
			errSlice.add(wrapPathErrorSlice(errs, propEnvironments)...)
		default:
			step, errs := visitStepNode(stepNode.key, stepNode.value, source)
			stage.Steps = append(stage.Steps, step)
			errSlice.add(wrapPathErrorSlice(errs, stepNode.key.value)...)
		}
	}
	return
}
