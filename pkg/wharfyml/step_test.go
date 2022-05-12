package wharfyml

import (
	"testing"

	"github.com/iver-wharf/wharf-cmd/internal/errtestutil"
	"github.com/iver-wharf/wharf-cmd/internal/yamltesting"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml/visit"
	"github.com/stretchr/testify/assert"
)

func TestVisitStep_ErrIfNotMap(t *testing.T) {
	key, node := yamltesting.NewKeyedNode(t, `myStep: 123`)
	_, errs := visitStepNode(key, node, Args{}, nil)
	errtestutil.RequireContainsErr(t, errs, visit.ErrInvalidFieldType)
}

func TestVisitStep_ErrIfEmpty(t *testing.T) {
	key, node := yamltesting.NewKeyedNode(t, `myStep: {}`)
	_, errs := visitStepNode(key, node, Args{}, nil)
	errtestutil.RequireContainsErr(t, errs, ErrStepEmpty)
}

func TestVisitStep_ErrIfMultipleStepTypes(t *testing.T) {
	key, node := yamltesting.NewKeyedNode(t, `
myStep:
  container: {}
  docker: {}
`)
	_, errs := visitStepNode(key, node, Args{}, nil)
	errtestutil.RequireContainsErr(t, errs, ErrStepMultipleStepTypes)
}

func TestVisitStep_Name(t *testing.T) {
	key, node := yamltesting.NewKeyedNode(t, `
myStep:
  helm-package: {}
`)
	step, errs := visitStepNode(key, node, Args{}, nil)
	if len(errs) > 0 {
		t.Logf("errs: %v", errs)
	}
	assert.Equal(t, "myStep", step.Name)
}
