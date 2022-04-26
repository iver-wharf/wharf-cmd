package varsub

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOSEnvSource(t *testing.T) {
	os.Clearenv()
	os.Setenv("MY_ENV_VAR", "should not find this value")
	os.Setenv("WHARF_VAR_MY_ENV_VAR", "my value")

	s := NewOSEnvSource("WHARF_VAR_")

	myEnvVar, ok := s.Lookup("MY_ENV_VAR")
	require.True(t, ok)
	assert.Equal(t, "my value", myEnvVar.Value)
}

func TestOSEnvSource_ListVars(t *testing.T) {
	os.Clearenv()
	os.Setenv("WHARF_VAR_MY_ENV_VAR", "my value")
	os.Setenv("WHARF_VAR_MY_OTHER_ENV_VAR", "my other value")

	s := NewOSEnvSource("WHARF_VAR_")

	envVars := s.ListVars()
	var got []string
	for _, envVar := range envVars {
		got = append(got, fmt.Sprintf("%s=%v", envVar.Key, envVar.Value))
	}

	want := []string{
		"MY_ENV_VAR=my value",
		"MY_OTHER_ENV_VAR=my other value",
	}
	assert.Equal(t, want, got)
}
