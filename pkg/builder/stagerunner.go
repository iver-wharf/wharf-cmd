package builder

import (
	"context"
	"errors"

	"github.com/iver-wharf/wharf-cmd/pkg/core/wharfyml"
)

type stageRunner struct {
	stepRun StepRunner
}

func (r stageRunner) RunStage(ctx context.Context, stage wharfyml.Stage) (StageResult, error) {
	// TODO: Run all steps in stage in parallell
	return StageResult{}, errors.New("not implemented")
}
