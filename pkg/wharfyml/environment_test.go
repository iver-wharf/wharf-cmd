package wharfyml

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDefEnvironments_ErrIfNotMap(t *testing.T) {
	_, errs := visitDocEnvironmentsNode(getNode(t, `123`))
	requireContainsErr(t, errs, ErrInvalidFieldType)
}

func TestParseDefEnvironments_ErrIfKeyNotString(t *testing.T) {
	_, errs := visitDocEnvironmentsNode(getNode(t, `123: {}`))
	requireContainsErr(t, errs, ErrKeyNotString)
}

func TestParseDefEnvironments_ErrIfKeyEmpty(t *testing.T) {
	_, errs := visitDocEnvironmentsNode(getNode(t, `"": {}`))
	requireContainsErr(t, errs, ErrKeyEmpty)
}

func TestParseDefEnvironments_ValidMapOfEnvs(t *testing.T) {
	envs, _ := visitDocEnvironmentsNode(getNode(t, `
myEnv1: {}
myEnv2: {}
myEnv3: {}
`))
	require.Len(t, envs, 3)
	_, hasEnv1 := envs["myEnv1"]
	_, hasEnv2 := envs["myEnv2"]
	_, hasEnv3 := envs["myEnv3"]
	assert.True(t, hasEnv1, "has myEnv1")
	assert.True(t, hasEnv2, "has myEnv2")
	assert.True(t, hasEnv3, "has myEnv3")
}

func TestParseEnvironment_SetsName(t *testing.T) {
	env, _ := visitEnvironmentNode(getKeyedNode(t, `myEnv: {}`))
	assert.Equal(t, "myEnv", env.Name)
}

func TestParseEnvironment_ErrIfEnvNotMap(t *testing.T) {
	_, errs := visitEnvironmentNode(getKeyedNode(t, `myEnv: 123`))
	requireContainsErr(t, errs, ErrInvalidFieldType)
}

func TestParseEnvironment_ErrIfInvalidVarType(t *testing.T) {
	_, errs := visitEnvironmentNode(getKeyedNode(t, `
myEnv:
  myVar: [123]
`))
	requireContainsErr(t, errs, ErrInvalidFieldType)
}

func TestParseEnvironment_ValidVarTypes(t *testing.T) {
	env, errs := visitEnvironmentNode(getKeyedNode(t, `
myEnv:
  myIntNeg: -123
  myIntPos: 123
  myFloat: 456.789
  myString: foo bar
  myBool: true
`))
	requireNotContainsErr(t, errs, ErrInvalidFieldType)
	want := map[string]interface{}{
		"myIntNeg": -123,
		"myIntPos": 123,
		"myFloat":  456.789,
		"myString": "foo bar",
		"myBool":   true,
	}
	assert.Equal(t, want, env.Vars)
}

func TestParseStageEnvironments_ErrIfNotArray(t *testing.T) {
	_, errs := visitStageEnvironmentsNode(getNode(t, `123`))
	requireContainsErr(t, errs, ErrInvalidFieldType)
}

func TestParseStageEnvironments_Valid(t *testing.T) {
	envs, errs := visitStageEnvironmentsNode(getNode(t, `[a, b, c]`))
	if len(errs) > 0 {
		t.Logf("errs: %v", errs)
	}
	want := []string{"a", "b", "c"}
	got := make([]string, len(envs))
	for i, env := range envs {
		got[i] = env.Name
	}
	assert.Equal(t, want, got)
}
