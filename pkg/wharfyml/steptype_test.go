package wharfyml

import (
	"testing"
)

func TestVisitStepType_ErrIfNotMap(t *testing.T) {
	key, node := getKeyedNode(t, `container: 123`)
	_, errs := visitStepTypeNode("", key, node, nil)
	requireContainsErr(t, errs, ErrInvalidFieldType)
}

func TestVisitStepType_ErrIfInvalidField(t *testing.T) {
	key, node := getKeyedNode(t, `
container:
  image: [123]`)
	_, errs := visitStepTypeNode("", key, node, nil)
	requireContainsErr(t, errs, ErrInvalidFieldType)
}

func TestVisitStepType_ErrIfMissingRequiredField(t *testing.T) {
	// in "container" step, "cmds" and "image" are required
	key, node := getKeyedNode(t, `container: {}`)
	_, errs := visitStepTypeNode("", key, node, nil)
	requireContainsErr(t, errs, ErrMissingRequired)
}
