package wharfyml

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TODO: Test the following:
// - error on invalid environment variable type

func TestParseStageEnvironments_ErrIfNotArray(t *testing.T) {
	_, errs := visitEnvironmentStringsNode(getNode(t, `123`))
	requireContainsErr(t, errs, ErrStageEnvsNotArray)
}

func TestParseStageEnvironments_Valid(t *testing.T) {
	envs, errs := visitEnvironmentStringsNode(getNode(t, `[a, b, c]`))
	if len(errs) > 0 {
		t.Logf("errs: %v", errs)
	}
	want := []string{"a", "b", "c"}
	assert.Equal(t, want, envs)
}
