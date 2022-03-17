package wharfyml

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVisitStep_ErrIfNotMap(t *testing.T) {
	key, node := getKeyedNode(t, `myStep: 123`)
	_, errs := visitStepNode(key, node, nil)
	requireContainsErr(t, errs, ErrInvalidFieldType)
}

func TestVisitStep_ErrIfEmpty(t *testing.T) {
	key, node := getKeyedNode(t, `myStep: {}`)
	_, errs := visitStepNode(key, node, nil)
	requireContainsErr(t, errs, ErrStepEmpty)
}

func TestVisitStep_ErrIfMultipleStepTypes(t *testing.T) {
	key, node := getKeyedNode(t, `
myStep:
  container: {}
  docker: {}
`)
	_, errs := visitStepNode(key, node, nil)
	requireContainsErr(t, errs, ErrStepMultipleStepTypes)
}

func TestVisitStep_Name(t *testing.T) {
	key, node := getKeyedNode(t, `
myStep:
  helm-package: {}
`)
	step, errs := visitStepNode(key, node, nil)
	if len(errs) > 0 {
		t.Logf("errs: %v", errs)
	}
	assert.Equal(t, "myStep", step.Name)
}
