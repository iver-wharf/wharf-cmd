package builder

import (
	"context"
	"strings"
	"time"

	"github.com/iver-wharf/wharf-cmd/pkg/core/wharfyml"
	"github.com/iver-wharf/wharf-core/pkg/logger"
)

var log = logger.New()

type Builder interface {
	Build(ctx context.Context, def wharfyml.BuildDefinition) (Result, error)
}

type StageRunner interface {
	RunStage(ctx context.Context, def wharfyml.Stage) StageResult
}

type StepRunner interface {
	RunStep(ctx context.Context, def wharfyml.Step) StepResult
}

type Result struct {
	Status      Status
	Environment string
	Stages      []StageResult
	Duration    time.Duration
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

type builder struct {
	stageRun StageRunner
}

func New(stepRun StepRunner) Builder {
	return builder{
		stageRun: NewStageRunner(stepRun),
	}
}

func (b builder) Build(ctx context.Context, def wharfyml.BuildDefinition) (Result, error) {
	var result Result
	start := time.Now()
	stagesCount := len(def.Stages)
	stagesDone := 0
	for _, stage := range def.Stages {
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
				WithString("failed", strings.Join(failed, ",")).
				WithString("cancelled", strings.Join(cancelled, ",")).
				Message("Failed stage. Skipping any further stages.")
			result.Status = res.Status
			break
		} else {
			log.Info().
				WithStringf("stages", "%d/%d", stagesDone, stagesCount).
				WithString("stage", res.Name).
				WithDuration("dur", res.Duration.Truncate(time.Second)).
				Message("Done with stage.")
		}
	}
	result.Duration = time.Since(start)
	return result, nil
}
