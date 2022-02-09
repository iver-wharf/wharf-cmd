package wharfyml

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseStep_ErrIfNotMap(t *testing.T) {
	_, errs := visitStepNode("myStep", getNode(t, `123`))
	requireContainsErr(t, errs, ErrNotMap)
}

func TestParseStep_ErrIfEmpty(t *testing.T) {
	_, errs := visitStepNode("myStep", getNode(t, `{}`))
	requireContainsErr(t, errs, ErrStepEmpty)
}

func TestParseStep_ErrIfMultipleStepTypes(t *testing.T) {
	_, errs := visitStepNode("myStep", getNode(t, `
container: {}
docker: {}
`))
	requireContainsErr(t, errs, ErrStepMultipleStepTypes)
}

func TestParseStep_Name(t *testing.T) {
	step, errs := visitStepNode("myStep", getNode(t, `
helm-package: {}
`))
	if len(errs) > 0 {
		t.Logf("errs: %v", errs)
	}
	assert.Equal(t, "myStep", step.Name)
}
