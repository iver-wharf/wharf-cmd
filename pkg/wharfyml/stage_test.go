package wharfyml

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseStage_ErrIfNotMap(t *testing.T) {
	_, errs := visitStageNode(getKeyedNode(t, "myStage", `123`))
	requireContainsErr(t, errs, ErrStageNotMap)
}

func TestParseStage_ErrIfEmptyKey(t *testing.T) {
	_, errs := visitStageNode(getKeyedNode(t, "", `{}`))
	requireContainsErr(t, errs, ErrStageEmptyName)
}

func TestParseStage_ErrIfEmptyMap(t *testing.T) {
	_, errs := visitStageNode(getKeyedNode(t, "myStage", `{}`))
	requireContainsErr(t, errs, ErrStageEmpty)
}

func TestParseStage_Name(t *testing.T) {
	stage, errs := visitStageNode(getKeyedNode(t, "myStage", `{}`))
	if len(errs) > 0 {
		t.Logf("errs: %v", errs)
	}
	assert.Equal(t, "myStage", stage.Name)
}
