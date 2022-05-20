package wharfyml

import (
	"testing"

	"github.com/iver-wharf/wharf-cmd/internal/testutil"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml/visit"
	"github.com/stretchr/testify/assert"
)

func TestVisitStep_ErrIfNotMap(t *testing.T) {
	key, node := testutil.NewKeyedNode(t, `myStep: 123`)
	_, errs := visitStepNode(key, node, Args{}, nil)
	testutil.RequireContainsErr(t, errs, visit.ErrInvalidFieldType)
}

func TestVisitStep_ErrIfEmpty(t *testing.T) {
	key, node := testutil.NewKeyedNode(t, `myStep: {}`)
	_, errs := visitStepNode(key, node, Args{}, nil)
	testutil.RequireContainsErr(t, errs, ErrStepEmpty)
}

func TestVisitStep_ErrIfMultipleStepTypes(t *testing.T) {
	key, node := testutil.NewKeyedNode(t, `
myStep:
  container: {}
  docker: {}
`)
	_, errs := visitStepNode(key, node, Args{}, nil)
	testutil.RequireContainsErr(t, errs, ErrStepMultipleStepTypes)
}

func TestVisitStep_Name(t *testing.T) {
	key, node := testutil.NewKeyedNode(t, `
myStep:
  helm-package: {}
`)
	step, errs := visitStepNode(key, node, Args{}, nil)
	if len(errs) > 0 {
		t.Logf("errs: %v", errs)
	}
	assert.Equal(t, "myStep", step.Name)
}
