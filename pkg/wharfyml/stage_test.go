package wharfyml

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVisitStage_ErrIfNotMap(t *testing.T) {
	key, node := getKeyedNode(t, `myStage: 123`)
	_, errs := visitStageNode(key, node, nil)
	requireContainsErr(t, errs, ErrInvalidFieldType)
}

func TestVisitStage_ErrIfEmptyMap(t *testing.T) {
	key, node := getKeyedNode(t, `myStage: {}`)
	_, errs := visitStageNode(key, node, nil)
	requireContainsErr(t, errs, ErrStageEmpty)
}

func TestVisitStage_Name(t *testing.T) {
	key, node := getKeyedNode(t, `myStage: {}`)
	stage, errs := visitStageNode(key, node, nil)
	if len(errs) > 0 {
		t.Logf("errs: %v", errs)
	}
	assert.Equal(t, "myStage", stage.Name)
}
