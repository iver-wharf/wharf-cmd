package wharfyml

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVisitStage_ErrIfNotMap(t *testing.T) {
	_, errs := visitStageNode(getKeyedNode(t, `myStage: 123`))
	requireContainsErr(t, errs, ErrInvalidFieldType)
}

func TestVisitStage_ErrIfEmptyMap(t *testing.T) {
	_, errs := visitStageNode(getKeyedNode(t, `myStage: {}`))
	requireContainsErr(t, errs, ErrStageEmpty)
}

func TestVisitStage_Name(t *testing.T) {
	stage, errs := visitStageNode(getKeyedNode(t, `myStage: {}`))
	if len(errs) > 0 {
		t.Logf("errs: %v", errs)
	}
	assert.Equal(t, "myStage", stage.Name)
}
