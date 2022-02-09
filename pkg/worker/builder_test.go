package worker

import (
	"container/ring"
	"context"
	"testing"

	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockStageRunner struct {
	results *ring.Ring
}

func (r *mockStageRunner) RunStage(_ context.Context, def wharfyml.Stage) StageResult {
	result := r.results.Value.(StageResult)
	r.results = r.results.Next()
	result.Name = def.Name
	return result
}

func TestBuilder_runsAllSuccess(t *testing.T) {
	stageRun := &mockStageRunner{results: ringOfStageResults(
		StageResult{Status: StatusSuccess},
	)}
	b := New(stageRun)
	def := wharfyml.BuildDefinition{
		Stages: map[string]wharfyml.Stage{
			"foo": {Name: "foo"},
			"bar": {Name: "bar"},
			"moo": {Name: "moo"},
		},
	}
	result, err := b.Build(context.Background(), def, BuildOptions{})
	require.NoError(t, err, "builder.Build")
	assert.Equal(t, StatusSuccess, result.Status, "result.Status")
	gotNames := getNamesFromStageResults(result.Stages)
	wantNames := []string{"foo", "bar", "moo"}
	assert.ElementsMatch(t, gotNames, wantNames, "result.Stages[].Name")
}

func TestBuilder_runsMiddleFails(t *testing.T) {
	stageRun := &mockStageRunner{results: ringOfStageResults(
		StageResult{Status: StatusSuccess},
		StageResult{Status: StatusFailed},
		StageResult{Status: StatusSuccess},
	)}
	b := New(stageRun)
	def := wharfyml.BuildDefinition{
		Stages: map[string]wharfyml.Stage{
			"foo": {Name: "foo"},
			"bar": {Name: "bar"},
			"moo": {Name: "moo"},
		},
	}
	result, err := b.Build(context.Background(), def, BuildOptions{})
	require.NoError(t, err, "builder.Build")
	assert.Equal(t, StatusFailed, result.Status, "result.Status")
	require.Len(t, result.Stages, 2, "result.Stages")
	assert.Equal(t, StatusSuccess, result.Stages[0].Status)
	assert.Equal(t, StatusFailed, result.Stages[1].Status)
}

func getNamesFromStageResults(stages []StageResult) []string {
	var names []string
	for _, stage := range stages {
		names = append(names, stage.Name)
	}
	return names
}

func ringOfStageResults(results ...StageResult) *ring.Ring {
	r := ring.New(len(results))
	for _, s := range results {
		r.Value = s
		r = r.Next()
	}
	return r
}
