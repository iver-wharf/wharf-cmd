package worker

import (
	"context"
	"time"

	"github.com/iver-wharf/wharf-cmd/pkg/core/wharfyml"
	"github.com/iver-wharf/wharf-core/pkg/logger"
)

// TODO: Change to factory pattern so the steps can get validated before
// it's their turn

var log = logger.New()

type BuildOptions struct {
	StageFilter string
}

type Builder interface {
	Build(ctx context.Context, def wharfyml.BuildDefinition, opt BuildOptions) (Result, error)
}

type StageRunner interface {
	RunStage(ctx context.Context, def wharfyml.Stage) StageResult
}

type StepRunner interface {
	RunStep(ctx context.Context, def wharfyml.Step) StepResult
}

type Result struct {
	Status   Status
	Options  BuildOptions
	Stages   []StageResult
	Duration time.Duration
}

type StageResult struct {
	Name     string
	Status   Status
	Steps    []StepResult
	Duration time.Duration
}

type StepResult struct {
	Name     string
	Status   Status
	Type     string
	Error    error
	Duration time.Duration
}
