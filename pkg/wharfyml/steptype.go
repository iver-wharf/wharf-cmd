package wharfyml

import (
	"errors"

	"github.com/iver-wharf/wharf-cmd/internal/errutil"
	"github.com/iver-wharf/wharf-cmd/pkg/varsub"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml/visit"
	"gopkg.in/yaml.v3"
)

// errutil.Slice related to parsing step types.
var (
	ErrNoStepTypesRegistered = errors.New("no step types registered")
)

// StepType is an interface that is implemented by all step types.
type StepType interface {
	StepTypeName() string
}

// StepTypeMeta contains metadata about a step type.
type StepTypeMeta struct {
	StepName string
	Source   visit.Pos
	FieldPos map[string]visit.Pos
}

type StepTypeFactory interface {
	NewStepType(stepTypeName, stepName string, v visit.MapVisitor) (StepType, errutil.Slice)
}

func visitStepTypeNode(stepName string, key visit.StringNode, node *yaml.Node, args Args, source varsub.Source) (StepType, StepTypeMeta, errutil.Slice) {
	var errSlice errutil.Slice
	m, errs := visit.Map(node)
	errSlice.Add(errs...)

	if args.StepTypeFactory == nil {
		errSlice.Add(ErrNoStepTypesRegistered)
		return nil, StepTypeMeta{}, errSlice
	}

	v := visit.NewMapVisitor(key.Node, m, source)
	stepType, errs := args.StepTypeFactory.NewStepType(key.Value, stepName, v)
	errSlice.Add(errs...)
	return stepType, getStepTypeMeta(v, stepName), errSlice
}

func getStepTypeMeta(v visit.MapVisitor, stepName string) StepTypeMeta {
	return StepTypeMeta{
		StepName: stepName,
		Source:   v.ParentPos(),
		FieldPos: v.ReadNodesPos(),
	}
}
