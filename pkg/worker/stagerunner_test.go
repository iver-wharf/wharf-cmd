package worker

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml"
	"github.com/iver-wharf/wharf-cmd/pkg/worker/workermodel"
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
			result.Status = workermodel.StatusCancelled
		case <-time.After(time.Second):
			result.Status = workermodel.StatusUnknown
		}
	}
	return result
}

func TestStageRunner_runAllSuccess(t *testing.T) {
	factory := mockStepRunFactory{runners: map[string]mockStepRunner{
		"foo": {result: StepResult{Status: workermodel.StatusSuccess}},
		"bar": {result: StepResult{Status: workermodel.StatusSuccess}},
		"moo": {result: StepResult{Status: workermodel.StatusSuccess}},
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
	assert.Equal(t, workermodel.StatusSuccess, result.Status)

	gotNames := getNamesFromStepResults(result.Steps)
	wantNames := []string{"foo", "bar", "moo"}
	assert.ElementsMatch(t, wantNames, gotNames)
}

func TestStageRunner_runOneFailsOthersCancelled(t *testing.T) {
	factory := mockStepRunFactory{runners: map[string]mockStepRunner{
		"foo": {result: StepResult{Status: workermodel.StatusFailed}, wait: true},
		"bar": {result: StepResult{Status: workermodel.StatusFailed}, wait: true},
		"moo": {result: StepResult{Status: workermodel.StatusFailed}, wait: false},
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
	assert.Equal(t, workermodel.StatusFailed, result.Status)

	gotNames := getNamesFromStepResults(result.Steps)
	wantNames := []string{"foo", "bar", "moo"}
	assert.ElementsMatch(t, wantNames, gotNames)

	gotStatuses := getStatusesFromStepResults(result.Steps)
	wantStatuses := map[string]workermodel.Status{
		"foo": workermodel.StatusCancelled,
		"bar": workermodel.StatusCancelled,
		"moo": workermodel.StatusFailed,
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

func getStatusesFromStepResults(steps []StepResult) map[string]workermodel.Status {
	statuses := make(map[string]workermodel.Status)
	for _, step := range steps {
		statuses[step.Name] = step.Status
	}
	return statuses
}
