package wharfyml

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseStage_ErrIfNotMap(t *testing.T) {
	_, errs := visitStageNode("myStage", getNode(t, `123`))
	requireContainsErr(t, errs, ErrNotMap)
}

func TestParseStage_ErrIfEmptyMap(t *testing.T) {
	_, errs := visitStageNode("myStage", getNode(t, `{}`))
	requireContainsErr(t, errs, ErrStageEmpty)
}

func TestParseStage_Name(t *testing.T) {
	stage, errs := visitStageNode("myStage", getNode(t, `{}`))
	if len(errs) > 0 {
		t.Logf("errs: %v", errs)
	}
	assert.Equal(t, "myStage", stage.Name)
}
