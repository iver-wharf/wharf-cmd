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
