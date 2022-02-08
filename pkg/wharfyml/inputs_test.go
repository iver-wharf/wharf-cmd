package wharfyml

import "testing"

func TestParseDefInputs_ErrIfNotArray(t *testing.T) {
	_, errs := visitDocInputsNode(getNode(t, `123`))
	requireContainsErr(t, errs, ErrInputsNotArray)
}

// TODO: more tests
