package wharfyml

import (
	"testing"
)

func TestParseStepType_ErrIfNotMap(t *testing.T) {
	_, errs := visitStepTypeNode(getKeyedNode(t, "container", `123`))
	requireContainsErr(t, errs, ErrStepTypeNotMap)
}

func TestParseStepType_ErrIfInvalidField(t *testing.T) {
	_, errs := visitStepTypeNode(getKeyedNode(t, "container", `image: [123]`))
	requireContainsErr(t, errs, ErrInvalidFieldType)
}

func TestParseStepType_ErrIfMissingRequiredField(t *testing.T) {
	// in "container" step, "cmds" and "image" are required
	_, errs := visitStepTypeNode(getKeyedNode(t, "container", `{}`))
	requireContainsErr(t, errs, ErrStepTypeMissingRequired)
}
