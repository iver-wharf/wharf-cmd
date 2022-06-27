package worker

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml"
	"github.com/iver-wharf/wharf-cmd/pkg/worker/workermodel"
)

type builder struct {
	opts         BuildOptions
	def          wharfyml.Definition
	stageRunners []StageRunner
}

// New returns a new Builder implementation that uses the provided StageRunner
// to run all build stages in series.
func New(ctx context.Context, stageRunFactory StageRunnerFactory, def wharfyml.Definition, opts BuildOptions) (Builder, error) {
	filteredStages := filterStages(def.Stages, opts.StageFilter)
	stageRunners := make([]StageRunner, len(filteredStages))
	stepIDOffset := uint64(1)
	for i, stage := range filteredStages {
		r, err := stageRunFactory.NewStageRunner(ctx, stage, stepIDOffset)
		stepIDOffset += uint64(len(stage.Steps))
		if err != nil {
			return nil, fmt.Errorf("stage %s: %w", stage.Name, err)
		}
		if r == nil {
			return nil, fmt.Errorf("stage %s: unexpected nil stage runner", stage.Name)
		}
		stageRunners[i] = r
	}
	return builder{
		stageRunners: stageRunners,
	}, nil
}

func (b builder) BuildOptions() BuildOptions {
	return b.opts
}

func (b builder) Definition() wharfyml.Definition {
	return b.def
}

func (b builder) Build(ctx context.Context) (Result, error) {
	var result Result
	start := time.Now()
	stagesCount := len(b.stageRunners)
	stagesDone := 0
	if stagesCount == 0 {
		log.Warn().
			WithString("stages", "0/0").
			Message("No stages to run.")
		result.Status = workermodel.StatusNone
		return result, nil
	}
	anyPreviousStageHasFailed := false
	for _, stageRunner := range b.stageRunners {
		stagesDone++
		stage := stageRunner.Stage()
		if stage.ShouldSkip(anyPreviousStageHasFailed) {
			logSkippedStage(stage, stagesDone, stagesCount)
			continue
		}
		log.Info().
			WithStringf("stages", "%d/%d", stagesDone, stagesCount).
			WithString("stage", stage.Name).
			Message("Starting stage.")
		res := stageRunner.RunStage(ctx)
		result.Stages = append(result.Stages, res)
		if res.Status != workermodel.StatusSuccess {
			logFailedStage(res, stagesDone, stagesCount)
			if !anyPreviousStageHasFailed {
				log.Debug().Message("Skipping `run-if: success` stages from now on.")
				anyPreviousStageHasFailed = true
			}
			result.Status = res.Status
			continue
		}
		logSuccessfulStage(res, stagesDone, stagesCount)
		result.Status = workermodel.StatusSuccess
	}
	if anyPreviousStageHasFailed {
		result.Status = workermodel.StatusFailed
	}
	if errors.Is(ctx.Err(), context.Canceled) {
		result.Status = workermodel.StatusCancelled
	}
	result.Duration = time.Since(start)
	return result, nil
}

func filterStages(stages []wharfyml.Stage, nameFilter string) []wharfyml.Stage {
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

func logSkippedStage(stage wharfyml.Stage, stagesDone, stagesCount int) {
	ev := log.Info().
		WithStringf("stages", "%d/%d", stagesDone, stagesCount).
		WithString("stage", stage.Name)
	if stage.RunsIf == wharfyml.StageRunsIfFail {
		ev = ev.WithString("reason", "only runs if any of the previous stages failed")
	} else {
		ev = ev.WithString("reason", "only runs if all previous stages succeeded")
	}
	ev.Message("Skipping stage.")
}

func logFailedStage(res StageResult, stagesDone, stagesCount int) {
	var failed []string
	var cancelled []string
	for _, stepRes := range res.Steps {
		if stepRes.Status == workermodel.StatusFailed {
			failed = append(failed, stepRes.Name)
		} else if stepRes.Status == workermodel.StatusCancelled {
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
		Message("Failed stage.")
}

func logSuccessfulStage(res StageResult, stagesDone, stagesCount int) {
	log.Info().
		WithStringf("stages", "%d/%d", stagesDone, stagesCount).
		WithString("stage", res.Name).
		WithDuration("dur", res.Duration.Truncate(time.Second)).
		Message("Done with stage.")
}
