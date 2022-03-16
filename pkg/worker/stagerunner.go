package worker

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml"
	"github.com/iver-wharf/wharf-cmd/pkg/worker/workermodel"
	"github.com/iver-wharf/wharf-core/pkg/logger"
)

// NewStageRunnerFactory returns a new StageRunner that uses the provided
// StepRunner to run the steps in parallel.
func NewStageRunnerFactory(stepRunFactory StepRunnerFactory) (StageRunnerFactory, error) {
	return stageRunnerFactory{stepRunFactory}, nil
}

type stageRunnerFactory struct {
	stepRunFactory StepRunnerFactory
}

// NewStageRunner returns a new StageRunner that uses the provided StepRunner to
// run the steps in parallel.
func (f stageRunnerFactory) NewStageRunner(ctx context.Context, stage wharfyml.Stage, stepIDOffset uint64) (StageRunner, error) {
	return newStageRunner(ctx, f.stepRunFactory, stage, stepIDOffset)
}

func newStageRunner(ctx context.Context, stepRunFactory StepRunnerFactory, stage wharfyml.Stage, stepIDOffset uint64) (StageRunner, error) {
	ctx = contextWithStageName(ctx, stage.Name)
	stepRunners := make([]StepRunner, len(stage.Steps))
	for i, step := range stage.Steps {
		r, err := stepRunFactory.NewStepRunner(ctx, step, stepIDOffset+uint64(i))
		if err != nil {
			return nil, fmt.Errorf("step %s: %w", step.Name, err)
		}
		stepRunners[i] = r
	}
	return stageRunner{stage, stepRunners}, nil
}

type stageRunner struct {
	stage       wharfyml.Stage
	stepRunners []StepRunner
}

func (r stageRunner) Stage() wharfyml.Stage {
	return r.stage
}

func (r stageRunner) RunStage(ctx context.Context) StageResult {
	ctx = contextWithStageName(ctx, r.stage.Name)
	stageRun := stageRun{
		stepCount: len(r.stepRunners),
		stage:     &r.stage,
		start:     time.Now(),
	}
	for _, stepRunner := range r.stepRunners {
		stageRun.startRunStepGoroutine(ctx, stepRunner)
	}
	return stageRun.waitForResult()
}

type stageRun struct {
	stage       *wharfyml.Stage
	cancelFuncs []func()
	stepCount   int
	stepsDone   int32

	failed      bool
	stepResults []StepResult
	start       time.Time

	wg    sync.WaitGroup
	mutex sync.Mutex
}

func (r *stageRun) startRunStepGoroutine(ctx context.Context, stepRunner StepRunner) {
	r.wg.Add(1)
	stepCtx, cancel := context.WithCancel(ctx)
	r.cancelFuncs = append(r.cancelFuncs, cancel)
	go r.runStep(stepCtx, stepRunner)
}

func (r *stageRun) waitForResult() StageResult {
	r.wg.Wait()
	status := workermodel.StatusSuccess
	if r.failed {
		status = workermodel.StatusFailed
	}
	return StageResult{
		Name:     r.stage.Name,
		Status:   status,
		Steps:    r.stepResults,
		Duration: time.Since(r.start),
	}
}

func (r *stageRun) addStepResult(res StepResult) {
	r.mutex.Lock()
	r.stepResults = append(r.stepResults, res)
	r.mutex.Unlock()
	atomic.AddInt32(&r.stepsDone, 1)
}

func (r *stageRun) runStep(ctx context.Context, stepRunner StepRunner) {
	defer r.wg.Done()
	logFunc := func(ev logger.Event) logger.Event {
		return ev.
			WithStringf("steps", "%d/%d", r.stepsDone, r.stepCount).
			WithString("stage", r.stage.Name).
			WithString("step", stepRunner.Step().Name)
	}
	log.Info().WithFunc(logFunc).Message("Starting step.")
	res := stepRunner.RunStep(ctx)
	r.addStepResult(res)
	dur := res.Duration.Truncate(time.Second)
	if res.Status == workermodel.StatusCancelled {
		log.Info().
			WithFunc(logFunc).
			WithDuration("dur", dur).
			Message("Cancelled pod.")
	} else if res.Status != workermodel.StatusSuccess {
		r.failed = true
		log.Warn().
			WithError(res.Error).
			WithFunc(logFunc).
			WithDuration("dur", dur).
			Message("Failed step. Cancelling other steps in stage.")
		for _, cancel := range r.cancelFuncs {
			cancel()
		}
	} else {
		log.Info().
			WithFunc(logFunc).
			WithDuration("dur", dur).
			Message("Done with step.")
	}
}
