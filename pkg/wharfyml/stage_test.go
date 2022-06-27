package wharfyml

import (
	"testing"

	"github.com/iver-wharf/wharf-cmd/internal/testutil"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml/visit"
	"github.com/stretchr/testify/assert"
)

func TestVisitStage_ErrIfNotMap(t *testing.T) {
	key, node := testutil.NewKeyedNode(t, `myStage: 123`)
	_, errs := visitStageNode(key, node, Args{}, nil)
	testutil.RequireContainsErr(t, errs, visit.ErrInvalidFieldType)
}

func TestVisitStage_ErrIfEmptyMap(t *testing.T) {
	key, node := testutil.NewKeyedNode(t, `myStage: {}`)
	_, errs := visitStageNode(key, node, Args{}, nil)
	testutil.RequireContainsErr(t, errs, ErrStageEmpty)
}

func TestVisitStage_Name(t *testing.T) {
	key, node := testutil.NewKeyedNode(t, `myStage: {}`)
	stage, errs := visitStageNode(key, node, Args{}, nil)
	if len(errs) > 0 {
		t.Logf("errs: %v", errs)
	}
	assert.Equal(t, "myStage", stage.Name)
}

func TestShouldRun(t *testing.T) {
	testCases := []struct {
		name                      string
		s                         Stage
		wantSkip                  bool
		anyPreviousStageHasFailed bool
	}{
		{
			name:                      "'always' should run on fail",
			s:                         Stage{RunsIf: StageRunsIfAlways},
			wantSkip:                  false,
			anyPreviousStageHasFailed: true,
		},
		{
			name:                      "'always' should run on success",
			s:                         Stage{RunsIf: StageRunsIfAlways},
			wantSkip:                  false,
			anyPreviousStageHasFailed: false,
		},
		{
			name:                      "'failed' should run on fail",
			s:                         Stage{RunsIf: StageRunsIfFail},
			wantSkip:                  false,
			anyPreviousStageHasFailed: true,
		},
		{
			name:                      "'failed' should not run on success",
			s:                         Stage{RunsIf: StageRunsIfFail},
			wantSkip:                  true,
			anyPreviousStageHasFailed: false,
		},
		{
			name:                      "'success' should run on success",
			s:                         Stage{RunsIf: StageRunsIfSuccess},
			wantSkip:                  false,
			anyPreviousStageHasFailed: false,
		},
		{
			name:                      "'success' should not run on fail",
			s:                         Stage{RunsIf: StageRunsIfSuccess},
			wantSkip:                  true,
			anyPreviousStageHasFailed: true,
		},
		{
			name:                      "empty should run on success - default value",
			s:                         Stage{},
			wantSkip:                  false,
			anyPreviousStageHasFailed: false,
		},
		{
			name:                      "empty should not run on fail - default value",
			s:                         Stage{},
			wantSkip:                  true,
			anyPreviousStageHasFailed: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.s.ShouldSkip(tc.anyPreviousStageHasFailed)
			assert.Equal(t, tc.wantSkip, got)
		})
	}
}
