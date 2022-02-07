package wharfyml

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TODO: Test the following:
// - error on invalid environment variable type

func TestParseStageEnvironments_ErrIfNotArray(t *testing.T) {
	_, errs := parse(strings.NewReader(`
myStage:
  environments: 123
`))
	requireContainsErr(t, errs, ErrStageEnvsNotArray)
}

func TestParseStageEnvironments_Valid(t *testing.T) {
	def, errs := parse(strings.NewReader(`
myStage:
  environments: [a, b, c]
`))
	if len(errs) > 0 {
		t.Logf("errs: %v", errs)
	}
	require.NotEmpty(t, def.Stages)
	want := []string{"a", "b", "c"}
	assert.Equal(t, want, def.Stages[0].Environments)
}
