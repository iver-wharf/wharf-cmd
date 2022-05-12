package wharfyml

import (
	"errors"

	"github.com/iver-wharf/wharf-cmd/internal/errutil"
	"github.com/iver-wharf/wharf-cmd/pkg/varsub"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml/visit"
	"gopkg.in/yaml.v3"
)

// Errors related to parsing stages.
var (
	ErrStageEmpty = errors.New("stage is missing steps")
)

// Stage holds the name, environment filter, and list of steps for this Wharf
// build stage.
type Stage struct {
	Pos     visit.Pos
	Name    string
	Envs    []EnvRef
	EnvsPos visit.Pos
	Steps   []Step
}

func visitStageNode(nameNode visit.StringNode, node *yaml.Node, args Args, source varsub.Source) (stage Stage, errSlice errutil.Slice) {
	stage.Pos = visit.NewPosNode(node)
	stage.Name = nameNode.Value
	nodes, errs := visit.MapSlice(node)
	errSlice.Add(errs...)
	if len(nodes) == 0 {
		errSlice.Add(visit.WrapPosErrorNode(ErrStageEmpty, node))
		return
	}
	for _, stepNode := range nodes {
		switch stepNode.Key.Value {
		case propEnvironments:
			stage.EnvsPos = visit.NewPosNode(stepNode.Value)
			envs, errs := visitStageEnvironmentsNode(stepNode.Value)
			stage.Envs = envs
			errSlice.Add(errutil.ScopeSlice(errs, propEnvironments)...)
		default:
			step, errs := visitStepNode(stepNode.Key, stepNode.Value, args, source)
			stage.Steps = append(stage.Steps, step)
			errSlice.Add(errutil.ScopeSlice(errs, stepNode.Key.Value)...)
		}
	}
	return
}
