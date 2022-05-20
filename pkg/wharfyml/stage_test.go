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
