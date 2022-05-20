package wharfyml

import (
	"testing"

	"github.com/iver-wharf/wharf-cmd/internal/testutil"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml/visit"
)

func TestVisitStepType_ErrIfNotMap(t *testing.T) {
	key, node := testutil.NewKeyedNode(t, `container: 123`)
	_, _, errs := visitStepTypeNode("", key, node, Args{}, testVarSource)
	testutil.RequireContainsErr(t, errs, visit.ErrInvalidFieldType)
}
