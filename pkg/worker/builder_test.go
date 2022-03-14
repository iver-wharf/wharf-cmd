package worker

import (
	"context"
	"fmt"
	"testing"

	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml"
	"github.com/iver-wharf/wharf-cmd/pkg/worker/workermodel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockStageRunFactory struct {
	runners map[string]mockStageRunner
}

func (f mockStageRunFactory) NewStageRunner(
	_ context.Context, stage wharfyml.Stage, _ int) (StageRunner, error) {
	runner, ok := f.runners[stage.Name]
	if !ok {
		return nil, fmt.Errorf("no stage runner found for %q", stage.Name)
	}
	runner.stage.Name = stage.Name
	runner.result.Name = stage.Name
	return runner, nil
}

type mockStageRunner struct {
	stage  wharfyml.Stage
	result StageResult
}

func (r mockStageRunner) Stage() wharfyml.Stage {
	return r.stage
}

func (r mockStageRunner) RunStage(context.Context) StageResult {
	return r.result
}

func TestBuilder_runsAllSuccess(t *testing.T) {
	factory := &mockStageRunFactory{runners: map[string]mockStageRunner{
		"foo": {result: StageResult{Status: workermodel.StatusSuccess}},
		"bar": {result: StageResult{Status: workermodel.StatusSuccess}},
		"moo": {result: StageResult{Status: workermodel.StatusSuccess}},
	}}
	def := wharfyml.Definition{
		Stages: []wharfyml.Stage{
			{Name: "foo"},
			{Name: "bar"},
			{Name: "moo"},
		},
	}
	b, err := New(context.Background(), factory, def, BuildOptions{})
	require.NoError(t, err)

	result, err := b.Build(context.Background())
	require.NoError(t, err, "builder.Build")
	assert.Equal(t, workermodel.StatusSuccess, result.Status, "result.Status")
	gotNames := getNamesFromStageResults(result.Stages)
	wantNames := []string{"foo", "bar", "moo"}
	assert.ElementsMatch(t, gotNames, wantNames, "result.Stages[].Name")
}

func TestBuilder_runsMiddleFails(t *testing.T) {
	factory := &mockStageRunFactory{runners: map[string]mockStageRunner{
		"foo": {result: StageResult{Status: workermodel.StatusSuccess}},
		"bar": {result: StageResult{Status: workermodel.StatusFailed}},
		"moo": {result: StageResult{Status: workermodel.StatusSuccess}},
	}}
	def := wharfyml.Definition{
		Stages: []wharfyml.Stage{
			{Name: "foo"},
			{Name: "bar"},
			{Name: "moo"},
		},
	}
	b, err := New(context.Background(), factory, def, BuildOptions{})
	require.NoError(t, err)

	result, err := b.Build(context.Background())
	require.NoError(t, err, "builder.Build")
	assert.Equal(t, workermodel.StatusFailed, result.Status, "result.Status")
	require.Len(t, result.Stages, 2, "result.Stages")
	assert.Equal(t, workermodel.StatusSuccess, result.Stages[0].Status)
	assert.Equal(t, workermodel.StatusFailed, result.Stages[1].Status)
}

func getNamesFromStageResults(stages []StageResult) []string {
	var names []string
	for _, stage := range stages {
		names = append(names, stage.Name)
	}
	return names
}
