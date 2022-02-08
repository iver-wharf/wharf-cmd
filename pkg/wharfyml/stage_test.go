package wharfyml

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseStageEnvironments(t *testing.T) {
	content := []interface{}{"dev", "stage", "prod"}

	environments, err := parseStageEnvironments(content)
	require.Nil(t, err)
	assert.ElementsMatch(t, content, environments)
}

func TestParseStageEnvironmentsFails(t *testing.T) {
	content := []interface{}{"dev", 1, "prod"}

	_, err := parseStageEnvironments(content)
	require.NotNil(t, err)
}

// -------------

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
