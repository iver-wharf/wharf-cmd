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

	RunsIf StageRunsIf

	Node visit.MapItem
}

// ShouldSkip returns true if the stage should be skipped based on its run
// conditions.
func (s Stage) ShouldSkip(anyPreviousStageHasFailed bool) bool {
	switch s.RunsIf {
	case StageRunsIfAlways:
		return false
	case StageRunsIfFail:
		return !anyPreviousStageHasFailed
	case "", StageRunsIfSuccess:
		return anyPreviousStageHasFailed
	}
	return true
}

func visitStageNode(nameNode visit.StringNode, node *yaml.Node, args Args, source varsub.Source) (Stage, errutil.Slice) {
	var errSlice errutil.Slice
	stage := Stage{
		Pos:  visit.NewPosFromNode(node),
		Name: nameNode.Value,
		Node: visit.MapItem{Key: nameNode, Value: node},
	}
	nodes, errs := visit.MapSlice(node)
	errSlice.Add(errs...)
	if len(nodes) == 0 {
		errSlice.Add(errutil.NewPosFromNode(ErrStageEmpty, node))
		return stage, errSlice
	}
	for _, stepNode := range nodes {
		switch stepNode.Key.Value {
		case propEnvironments:
			stage.EnvsPos = visit.NewPosFromNode(stepNode.Value)
			envs, errs := visitStageEnvironmentsNode(stepNode.Value)
			stage.Envs = envs
			errSlice.Add(errutil.ScopeSlice(errs, propEnvironments)...)
		case propRunsIf:
			runsIf, errs := visitStageRunsIfNode(stepNode.Value)
			stage.RunsIf = runsIf
			errSlice.Add(errutil.ScopeSlice(errs, propRunsIf)...)
		default:
			step, errs := visitStepNode(stepNode.Key, stepNode.Value, args, source)
			stage.Steps = append(stage.Steps, step)
			errSlice.Add(errutil.ScopeSlice(errs, stepNode.Key.Value)...)
		}
	}
	return stage, errSlice
}
