package worker

import (
	"context"
	"time"

	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml"
	"github.com/iver-wharf/wharf-cmd/pkg/worker/workermodel"
	"github.com/iver-wharf/wharf-core/pkg/logger"
)

var log = logger.New()

// BuildOptions defines filtering options to control what parts of a build should
// actually be executed.
type BuildOptions struct {
	StageFilter string
}

// Builder is the interface for running a Wharf build. A single Wharf build may
// contain any number of stages, which in turn may contain any number of steps.
// All stages will be run in sequence.
type Builder interface {
	Definition() wharfyml.Definition
	Build(ctx context.Context) (Result, error)
}

// StageRunner is the interface for running a Wharf build stage. A single Wharf
// build stage may contain any number of steps which will all be run in
// parallel.
type StageRunner interface {
	Stage() wharfyml.Stage
	RunStage(ctx context.Context) StageResult
}

// StageRunnerFactory creates a new StageRunner for a given stage.
type StageRunnerFactory interface {
	NewStageRunner(ctx context.Context, stage wharfyml.Stage, startStepID uint64) (StageRunner, error)
}

// StepRunner is the interface for running a Wharf build step. Steps are the
// smallest unit of work in Wharf, and each step represents a single Kubernetes
// pod or Docker container.
type StepRunner interface {
	Step() wharfyml.Step
	RunStep(ctx context.Context) StepResult
}

// StepRunnerFactory creates a new StepRunner for a given step.
type StepRunnerFactory interface {
	NewStepRunner(ctx context.Context, step wharfyml.Step, stepID uint64) (StepRunner, error)
}

// Result is a Wharf build result with the overall status of all stages were
// executed, the induvidual stage results, as well as the duration of the entire
// Wharf build.
type Result struct {
	Status   workermodel.Status // execution status of the entire build
	Options  BuildOptions       // options used when running the build
	Stages   []StageResult      // execution results for each stage
	Duration time.Duration      // execution duration of the entire build
}

// StageResult is a Wharf build stage result with the overall status of the
// steps that was executed for the stage, as well as the duration of the
// Wharf build stage.
type StageResult struct {
	Name     string             // name of the stage
	Status   workermodel.Status // execution status of the stage
	Steps    []StepResult       // execution results for each step
	Duration time.Duration      // execution duration of the stage
}

// StepResult is a Wharf build step result with the status of the step execution
// as well as the duration of the Wharf build step.
type StepResult struct {
	Name     string             // name of the step
	Status   workermodel.Status // execution status of the step
	Type     string             // type of Wharf build step, eg. "container" or "docker"
	Error    error              // error message from the execution, if any
	Duration time.Duration      // execution duration of the step
}
