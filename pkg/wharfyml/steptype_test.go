package wharfyml

import (
	"testing"
)

func TestVisitStepType_ErrIfNotMap(t *testing.T) {
	_, errs := visitStepTypeNode(getKeyedNode(t, `container: 123`))
	requireContainsErr(t, errs, ErrInvalidFieldType)
}

func TestVisitStepType_ErrIfInvalidField(t *testing.T) {
	_, errs := visitStepTypeNode(getKeyedNode(t, `
container:
  image: [123]`))
	requireContainsErr(t, errs, ErrInvalidFieldType)
}

func TestVisitStepType_ErrIfMissingRequiredField(t *testing.T) {
	// in "container" step, "cmds" and "image" are required
	_, errs := visitStepTypeNode(getKeyedNode(t, `container: {}`))
	requireContainsErr(t, errs, ErrMissingRequired)
}
