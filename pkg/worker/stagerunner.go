package worker

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml"
	"github.com/iver-wharf/wharf-core/pkg/logger"
)

// NewStageRunner returns a new StageRunner that uses the provided StepRunner to
// run the steps in parallel.
func NewStageRunner(stepRun StepRunner) StageRunner {
	return stageRunner{stepRun}
}

type stageRunner struct {
	stepRun StepRunner
}

func (r stageRunner) RunStage(ctx context.Context, stage wharfyml.Stage) StageResult {
	ctx = contextWithStageName(ctx, stage.Name)
	stageRun := stageRun{
		stepRun:   r.stepRun,
		stepCount: len(stage.Steps),
		stage:     &stage,
		start:     time.Now(),
	}
	for _, step := range stage.Steps {
		stageRun.startRunStepGoroutine(ctx, step)
	}
	return stageRun.waitForResult()
}

type stageRun struct {
	stage       *wharfyml.Stage
	stepRun     StepRunner
	cancelFuncs []func()
	stepCount   int
	stepsDone   int32

	failed      bool
	stepResults []StepResult
	start       time.Time

	wg    sync.WaitGroup
	mutex sync.Mutex
}

func (r *stageRun) startRunStepGoroutine(ctx context.Context, step wharfyml.Step) {
	r.wg.Add(1)
	stepCtx, cancel := context.WithCancel(ctx)
	r.cancelFuncs = append(r.cancelFuncs, cancel)
	go r.runStep(stepCtx, step)
}

func (r *stageRun) waitForResult() StageResult {
	r.wg.Wait()
	status := StatusSuccess
	if r.failed {
		status = StatusFailed
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

func (r *stageRun) runStep(ctx context.Context, step wharfyml.Step) {
	defer r.wg.Done()
	logFunc := func(ev logger.Event) logger.Event {
		return ev.
			WithStringf("steps", "%d/%d", r.stepsDone, r.stepCount).
			WithString("stage", r.stage.Name).
			WithString("step", step.Name)
	}
	log.Info().WithFunc(logFunc).Message("Starting step.")
	res := r.stepRun.RunStep(ctx, step)
	r.addStepResult(res)
	dur := res.Duration.Truncate(time.Second)
	if res.Status == StatusCancelled {
		log.Info().
			WithFunc(logFunc).
			WithDuration("dur", dur).
			Message("Cancelled pod.")
	} else if res.Status != StatusSuccess {
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
