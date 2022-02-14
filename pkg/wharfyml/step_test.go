package wharfyml

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVisitStep_ErrIfNotMap(t *testing.T) {
	_, errs := visitStepNode(getKeyedNode(t, `myStep: 123`))
	requireContainsErr(t, errs, ErrInvalidFieldType)
}

func TestVisitStep_ErrIfEmpty(t *testing.T) {
	_, errs := visitStepNode(getKeyedNode(t, `myStep: {}`))
	requireContainsErr(t, errs, ErrStepEmpty)
}

func TestVisitStep_ErrIfMultipleStepTypes(t *testing.T) {
	_, errs := visitStepNode(getKeyedNode(t, `
myStep:
  container: {}
  docker: {}
`))
	requireContainsErr(t, errs, ErrStepMultipleStepTypes)
}

func TestVisitStep_Name(t *testing.T) {
	step, errs := visitStepNode(getKeyedNode(t, `
myStep:
  helm-package: {}
`))
	if len(errs) > 0 {
		t.Logf("errs: %v", errs)
	}
	assert.Equal(t, "myStep", step.Name)
}
