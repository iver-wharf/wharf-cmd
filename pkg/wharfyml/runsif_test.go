package wharfyml

import (
	"testing"

	"github.com/iver-wharf/wharf-cmd/internal/testutil"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml/visit"
)

func TestVisitRunsIf_ErrIfNotString(t *testing.T) {
	node := testutil.NewNode(t, "123")
	_, errs := visitStageRunsIfNode(node)
	testutil.RequireContainsErr(t, errs, visit.ErrInvalidFieldType)
}

func TestVisitRunsIf_ErrIfNotValidValue(t *testing.T) {
	node := testutil.NewNode(t, "badvalue")
	_, errs := visitStageRunsIfNode(node)
	testutil.RequireContainsErr(t, errs, ErrInvalidRunCondition)
}

func TestVisitRunsIf_Success(t *testing.T) {
	testCases := []string{"success", "always", "fail"}
	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			node := testutil.NewNode(t, tc)
			_, errs := visitStageRunsIfNode(node)
			testutil.RequireNoErr(t, errs)
		})
	}
}
