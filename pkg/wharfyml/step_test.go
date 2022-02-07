package wharfyml

import (
	"strings"
	"testing"
)

func TestParseStep_ErrIfNotMap(t *testing.T) {
	_, errs := parse(strings.NewReader(`
myStage:
  myStep: 123
`))
	requireContainsErr(t, errs, ErrStepNotMap)
}

func TestParseStep_ErrIfEmpty(t *testing.T) {
	_, errs := parse(strings.NewReader(`
myStage:
  myStep: {}
`))
	requireContainsErr(t, errs, ErrStepEmpty)
}

func TestParseStep_ErrIfMultipleStepTypes(t *testing.T) {
	_, errs := parse(strings.NewReader(`
myStage:
  myStep:
    container: {}
    docker: {}
`))
	requireContainsErr(t, errs, ErrStepMultipleStepTypes)
}

func TestParseStep_ErrIfInvalidField(t *testing.T) {
	_, errs := parse(strings.NewReader(`
myStage:
  myStep:
    container:
      image: 123
`))
	requireContainsErr(t, errs, ErrStepTypeInvalidField)
}

func TestParseStep_ErrIfMissingRequiredField(t *testing.T) {
	_, errs := parse(strings.NewReader(`
myStage:
  myStep:
    container: {} # cmds and image are required
`))
	requireContainsErr(t, errs, ErrStepTypeMissingRequired)
}
