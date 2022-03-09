package worker

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockStepRunFactory struct {
	runners map[string]mockStepRunner
}

func (f mockStepRunFactory) NewStepRunner(
	_ context.Context, step wharfyml.Step) (StepRunner, error) {
	runner, ok := f.runners[step.Name]
	if !ok {
		return nil, fmt.Errorf("no step runner found for %q", step.Name)
	}
	runner.step.Name = step.Name
	runner.result.Name = step.Name
	return runner, nil
}

type mockStepRunner struct {
	step   wharfyml.Step
	result StepResult
	wait   bool
}

func (r mockStepRunner) Step() wharfyml.Step {
	return r.step
}

func (r mockStepRunner) RunStep(ctx context.Context) StepResult {
	result := r.result
	if r.wait {
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
	factory := mockStepRunFactory{runners: map[string]mockStepRunner{
		"foo": {result: StepResult{Status: StatusSuccess}},
		"bar": {result: StepResult{Status: StatusSuccess}},
		"moo": {result: StepResult{Status: StatusSuccess}},
	}}
	stage := wharfyml.Stage{
		Name: "doesnt-matter",
		Steps: []wharfyml.Step{
			{Name: "foo"},
			{Name: "bar"},
			{Name: "moo"},
		},
	}
	b, err := newStageRunner(context.Background(), factory, stage)
	require.NoError(t, err)
	result := b.RunStage(context.Background())
	assert.Equal(t, StatusSuccess, result.Status)

	gotNames := getNamesFromStepResults(result.Steps)
	wantNames := []string{"foo", "bar", "moo"}
	assert.ElementsMatch(t, wantNames, gotNames)
}

func TestStageRunner_runOneFailsOthersCancelled(t *testing.T) {
	factory := mockStepRunFactory{runners: map[string]mockStepRunner{
		"foo": {result: StepResult{Status: StatusFailed}, wait: true},
		"bar": {result: StepResult{Status: StatusFailed}, wait: true},
		"moo": {result: StepResult{Status: StatusFailed}, wait: false},
	}}
	stage := wharfyml.Stage{
		Name: "doesnt-matter",
		Steps: []wharfyml.Step{
			{Name: "foo"},
			{Name: "bar"},
			{Name: "moo"},
		},
	}
	b, err := newStageRunner(context.Background(), factory, stage)
	require.NoError(t, err)
	result := b.RunStage(context.Background())
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
