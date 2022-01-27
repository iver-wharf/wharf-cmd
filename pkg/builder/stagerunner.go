package builder

import (
	"context"
	"sync"
	"time"

	"github.com/iver-wharf/wharf-cmd/pkg/core/wharfyml"
	"github.com/iver-wharf/wharf-core/pkg/logger"
)

func NewStageRunner(stepRun StepRunner) StageRunner {
	return stageRunner{stepRun}
}

type stageRunner struct {
	stepRun StepRunner
}

func (r stageRunner) RunStage(ctx context.Context, stage wharfyml.Stage) (StageResult, error) {
	ctx = contextWithStageName(ctx, stage.Name)
	start := time.Now()
	result := StageResult{
		Name:    stage.Name,
		Steps:   make([]StepResult, len(stage.Steps)),
		Success: true,
	}
	var wg sync.WaitGroup
	for i, step := range stage.Steps {
		wg.Add(1)
		go func(i int, step wharfyml.Step) {
			defer wg.Done()
			logFunc := func(ev logger.Event) logger.Event {
				return ev.WithString("stage", stage.Name).
					WithString("step", step.Name)
			}
			log.Info().WithFunc(logFunc).Message("Starting step.")
			res := r.stepRun.RunStep(ctx, step)
			result.Steps[i] = res
			if !res.Success {
				result.Success = false
				log.Warn().WithFunc(logFunc).WithError(res.Error).Message("Failed step.")
				// TODO: cancel all steps in stage via context
			} else {
				log.Info().WithFunc(logFunc).Message("Done with step.")
			}
		}(i, step)
	}
	wg.Wait()
	result.Duration = time.Since(start)
	return result, nil
}
