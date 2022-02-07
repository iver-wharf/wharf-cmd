package wharfyml

import (
	"strings"
	"testing"
)

func TestParseStepType_ErrIfInvalidField(t *testing.T) {
	_, errs := Parse2(strings.NewReader(`
myStage:
  myStep:
    container:
      image: []
`))
	requireContainsErr(t, errs, ErrStepTypeInvalidField)
}

func TestParseStepType_ErrIfMissingRequiredField(t *testing.T) {
	_, errs := Parse2(strings.NewReader(`
myStage:
  myStep:
    container: {} # cmds and image are required
`))
	requireContainsErr(t, errs, ErrStepTypeMissingRequired)
}
