package wharfyml

import (
	"errors"

	"github.com/iver-wharf/wharf-cmd/internal/errutil"
	"github.com/iver-wharf/wharf-cmd/pkg/varsub"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml/visit"
	"gopkg.in/yaml.v3"
)

// Errors related to parsing steps.
var (
	ErrStepEmpty             = errors.New("missing a step type")
	ErrStepMultipleStepTypes = errors.New("contains multiple step types")
)

// Step holds the step type and name of this Wharf build step.
type Step struct {
	Pos  visit.Pos
	Name string
	Type StepType
	Meta StepTypeMeta
}

func visitStepNode(name visit.StringNode, node *yaml.Node, args Args, source varsub.Source) (step Step, errSlice errutil.Slice) {
	step.Pos = visit.NewPosFromNode(node)
	step.Name = name.Value
	nodes, errs := visit.MapSlice(node)
	errSlice.Add(errs...)
	if len(nodes) == 0 {
		errSlice.Add(errutil.NewPosFromNode(ErrStepEmpty, node))
		return
	}
	if len(nodes) > 1 {
		errSlice.Add(errutil.NewPosFromNode(ErrStepMultipleStepTypes, node))
		// Continue, its not a fatal issue
	}
	for _, stepTypeNode := range nodes {
		stepType, meta, errs := visitStepTypeNode(
			name.Value, stepTypeNode.Key, stepTypeNode.Value, args, source)
		step.Type = stepType
		step.Meta = meta
		if stepType != nil {
			errSlice.Add(errutil.ScopeSlice(errs, stepType.StepTypeName())...)
		} else {
			errSlice.Add(errs...)
		}
	}
	return
}
