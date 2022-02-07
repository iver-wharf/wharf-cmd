package wharfyml

import (
	"strings"
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
	_, errs := parse(strings.NewReader(`
myStage: 123
`))
	requireContainsErr(t, errs, ErrStageNotMap)
}

func TestParseStage_ErrIfEmptyKey(t *testing.T) {
	_, errs := parse(strings.NewReader(`
"": {}
`))
	requireContainsErr(t, errs, ErrStageEmptyName)
}

func TestParseStage_ErrIfEmptyMap(t *testing.T) {
	_, errs := parse(strings.NewReader(`
myStage: {}
`))
	requireContainsErr(t, errs, ErrStageEmpty)
}

func TestParseStage_Name(t *testing.T) {
	def, errs := parse(strings.NewReader(`
myStage: {}
`))
	if len(errs) > 0 {
		t.Logf("errs: %v", errs)
	}
	require.NotEmpty(t, def.Stages)
	assert.Equal(t, "myStage", def.Stages[0].Name)
}
