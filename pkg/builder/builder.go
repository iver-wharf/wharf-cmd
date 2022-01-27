package builder

import (
	"errors"

	"github.com/iver-wharf/wharf-cmd/pkg/core/wharfyml"
	"github.com/iver-wharf/wharf-core/pkg/logger"
)

var log = logger.New()

type Builder interface {
	Build(def wharfyml.BuildDefinition) (Result, error)
}

type StageRunner interface {
	RunStage(def wharfyml.Stage) (StageResult, error)
}

type StepRunner interface {
	RunStep(def wharfyml.Step) StepResult
}

type Result struct {
	Environment string
	Success     bool
	Stages      []StageResult
}

type StageResult struct {
	Name    string
	Success bool
	Steps   []StepResult
}

type StepResult struct {
	Name  string
	Type  string
	Error error
}

type builder struct {
	stageRun StageRunner
}

func New(stepRun StepRunner) Builder {
	return builder{
		stageRun: stageRunner{stepRun},
	}
}

func (b builder) Build(def wharfyml.BuildDefinition) (Result, error) {
	// TODO: Run all stages in def in series
	return Result{}, errors.New("not implemented")
}

type stageRunner struct {
	stepRun StepRunner
}

func (r stageRunner) RunStage(stage wharfyml.Stage) (StageResult, error) {
	// TODO: Run all steps in stage in parallell
	return StageResult{}, errors.New("not implemented")
}
