package wharfyml

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDefEnvironments_ErrIfNotMap(t *testing.T) {
	_, errs := visitDocEnvironmentsNode(getNode(t, `123`))
	requireContainsErr(t, errs, ErrNotMap)
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
	env, _ := visitEnvironmentNode("myEnv", getNode(t, `{}`))
	assert.Equal(t, "myEnv", env.Name)
}

func TestParseEnvironment_ErrIfEnvNotMap(t *testing.T) {
	_, errs := visitEnvironmentNode("myEnv", getNode(t, `123`))
	requireContainsErr(t, errs, ErrNotMap)
}

func TestParseEnvironment_ErrIfInvalidVarType(t *testing.T) {
	_, errs := visitEnvironmentNode("myEnv", getNode(t, `
myVar: [123]
`))
	requireContainsErr(t, errs, ErrEnvInvalidVarType)
}

func TestParseEnvironment_ValidVarTypes(t *testing.T) {
	env, errs := visitEnvironmentNode("myEnv", getNode(t, `
myInt: -123
myUint: 123
myFloat: 456.789
myString: foo bar
myBool: true
`))
	requireNotContainsErr(t, errs, ErrEnvInvalidVarType)
	want := map[string]interface{}{
		"myInt":    int64(-123),
		"myUint":   uint64(123),
		"myFloat":  456.789,
		"myString": "foo bar",
		"myBool":   true,
	}
	assert.Equal(t, want, env.Vars)
}

func TestParseStageEnvironments_ErrIfNotArray(t *testing.T) {
	_, errs := visitStageEnvironmentsNode(getNode(t, `123`))
	requireContainsErr(t, errs, ErrStageEnvsNotArray)
}

func TestParseStageEnvironments_Valid(t *testing.T) {
	envs, errs := visitStageEnvironmentsNode(getNode(t, `[a, b, c]`))
	if len(errs) > 0 {
		t.Logf("errs: %v", errs)
	}
	want := []string{"a", "b", "c"}
	assert.Equal(t, want, envs)
}
