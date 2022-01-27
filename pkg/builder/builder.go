package builder

import (
	"context"
	"errors"
	"time"

	"github.com/iver-wharf/wharf-cmd/pkg/core/wharfyml"
	"github.com/iver-wharf/wharf-core/pkg/logger"
)

var log = logger.New()

type Builder interface {
	Build(ctx context.Context, def wharfyml.BuildDefinition) (Result, error)
}

type StageRunner interface {
	RunStage(ctx context.Context, def wharfyml.Stage) (StageResult, error)
}

type StepRunner interface {
	RunStep(ctx context.Context, def wharfyml.Step) StepResult
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
	Name     string
	Type     string
	Success  bool
	Error    error
	Duration time.Duration
}

type builder struct {
	stageRun StageRunner
}

func New(stepRun StepRunner) Builder {
	return builder{
		stageRun: stageRunner{stepRun},
	}
}

func (b builder) Build(ctx context.Context, def wharfyml.BuildDefinition) (Result, error) {
	// TODO: Run all stages in def in series
	return Result{}, errors.New("not implemented")
}
