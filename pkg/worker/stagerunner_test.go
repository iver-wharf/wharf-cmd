package worker

import (
	"context"
	"testing"
	"time"

	"github.com/iver-wharf/wharf-cmd/pkg/core/wharfyml"
	"github.com/stretchr/testify/assert"
)

type mockStepRunner struct {
	results map[string]StepResult
	waits   map[string]bool
}

func (r *mockStepRunner) RunStep(ctx context.Context, def wharfyml.Step) StepResult {
	result := r.results[def.Name]
	result.Name = def.Name
	if r.waits[def.Name] {
		select {
		case <-ctx.Done():
			result.Status = StatusCancelled
		case <-time.After(time.Second):
			result.Status = StatusUnknown
		}
	}
	return result
}

func TestStageRunner_runAllSuccess(t *testing.T) {
	stepRun := &mockStepRunner{results: map[string]StepResult{
		"foo": {Status: StatusSuccess},
		"bar": {Status: StatusSuccess},
		"moo": {Status: StatusSuccess},
	}}
	b := NewStageRunner(stepRun)
	def := wharfyml.Stage{
		Name: "doesnt-matter",
		Steps: []wharfyml.Step{
			{Name: "foo"},
			{Name: "bar"},
			{Name: "moo"},
		},
	}
	result := b.RunStage(context.Background(), def)
	assert.Equal(t, StatusSuccess, result.Status)
	gotNames := getNamesFromStepResults(result.Steps)
	wantNames := []string{"foo", "bar", "moo"}
	assert.ElementsMatch(t, wantNames, gotNames)
}

func TestStageRunner_runOneFailsOthersCancelled(t *testing.T) {
	stepRun := &mockStepRunner{
		results: map[string]StepResult{
			"foo": {Status: StatusFailed},
			"bar": {Status: StatusFailed},
			"moo": {Status: StatusFailed},
		},
		waits: map[string]bool{
			"foo": true,
			"bar": true,
			"moo": false,
		},
	}
	b := NewStageRunner(stepRun)
	def := wharfyml.Stage{
		Name: "doesnt-matter",
		Steps: []wharfyml.Step{
			{Name: "foo"},
			{Name: "bar"},
			{Name: "moo"},
		},
	}
	result := b.RunStage(context.Background(), def)
	assert.Equal(t, StatusFailed, result.Status)

	gotNames := getNamesFromStepResults(result.Steps)
	wantNames := []string{"foo", "bar", "moo"}
	assert.ElementsMatch(t, wantNames, gotNames)

	gotStatuses := getStatusesFromStepResults(result.Steps)
	wantStatuses := map[string]Status{
		"foo": StatusCancelled,
		"bar": StatusCancelled,
		"moo": StatusFailed,
	}
	assert.Equal(t, wantStatuses, gotStatuses)
}

func getNamesFromStepResults(steps []StepResult) []string {
	var names []string
	for _, step := range steps {
		names = append(names, step.Name)
	}
	return names
}

func getStatusesFromStepResults(steps []StepResult) map[string]Status {
	statuses := make(map[string]Status)
	for _, step := range steps {
		statuses[step.Name] = step.Status
	}
	return statuses
}
