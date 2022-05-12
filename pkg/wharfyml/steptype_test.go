package wharfyml

import (
	"testing"

	"github.com/iver-wharf/wharf-cmd/internal/errtesting"
	"github.com/iver-wharf/wharf-cmd/internal/yamltesting"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml/visit"
)

func TestVisitStepType_ErrIfNotMap(t *testing.T) {
	key, node := yamltesting.NewKeyedNode(t, `container: 123`)
	_, _, errs := visitStepTypeNode("", key, node, Args{}, testVarSource)
	errtesting.RequireContainsErr(t, errs, visit.ErrInvalidFieldType)
}
