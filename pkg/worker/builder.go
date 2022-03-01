package worker

import (
	"context"
	"strings"
	"time"

	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml"
)

type builder struct {
	stageRun StageRunner
	steps    []wharfyml.Step
}

// New returns a new Builder implementation that uses the provided StageRunner
// to run all build stages in series.
func New(stageRun StageRunner) Builder {
	return builder{
		stageRun: stageRun,
	}
}

func (b builder) Build(ctx context.Context, def wharfyml.Definition, opt BuildOptions) (Result, error) {
	result := Result{Options: opt}
	start := time.Now()
	stages := b.filterStages(def.Stages, opt.StageFilter)
	stagesCount := len(stages)
	stagesDone := 0
	if stagesCount == 0 {
		log.Warn().
			WithString("stages", "0/0").
			Message("No stages to run.")
		result.Status = StatusNone
		return result, nil
	}
	for _, stage := range stages {
		b.steps = append(b.steps, stage.Steps...)
	}
	for _, stage := range stages {
		log.Info().
			WithStringf("stages", "%d/%d", stagesDone, stagesCount).
			WithString("stage", stage.Name).
			Message("Starting stage.")
		res := b.stageRun.RunStage(ctx, stage)
		result.Stages = append(result.Stages, res)
		stagesDone++
		if res.Status != StatusSuccess {
			var failed []string
			var cancelled []string
			for _, stepRes := range res.Steps {
				if stepRes.Status == StatusFailed {
					failed = append(failed, stepRes.Name)
				} else if stepRes.Status == StatusCancelled {
					cancelled = append(cancelled, stepRes.Name)
				}
			}
			log.Warn().
				WithStringf("stages", "%d/%d", stagesDone, stagesCount).
				WithString("stage", res.Name).
				WithDuration("dur", res.Duration.Truncate(time.Second)).
				WithStringer("status", res.Status).
				WithString("failed", strings.Join(failed, ",")).
				WithString("cancelled", strings.Join(cancelled, ",")).
				Message("Failed stage. Skipping any further stages.")
			result.Status = res.Status
			break
		}
		log.Info().
			WithStringf("stages", "%d/%d", stagesDone, stagesCount).
			WithString("stage", res.Name).
			WithDuration("dur", res.Duration.Truncate(time.Second)).
			Message("Done with stage.")
		result.Status = StatusSuccess
	}
	result.Duration = time.Since(start)
	return result, nil
}

func (b builder) ListBuildSteps() []wharfyml.Step {
	return b.steps
}

func (b builder) filterStages(stages []wharfyml.Stage, nameFilter string) []wharfyml.Stage {
	var result []wharfyml.Stage
	for _, stage := range stages {
		if nameFilter == "" || stage.Name == nameFilter {
			result = append(result, stage)
		} else {
			log.Debug().
				WithString("stage", stage.Name).
				WithString("filter", nameFilter).
				Message("Skipping stage because of filter.")
		}
	}
	return result
}
