package wharfyml

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func requireContainsErr(t *testing.T, errs Errors, err error) {
	for _, e := range errs {
		if errors.Is(e, err) {
			return
		}
	}
	t.Fatalf("\nexpected contains error: %q\nactual: (len=%d)\n%s",
		err, len(errs), formatErrorSlice("  - ", errs))
}

func requireNotContainsErr(t *testing.T, errs Errors, err error) {
	for i, e := range errs {
		if errors.Is(e, err) {
			t.Fatalf("\nexpected not to contain error: %q\nfound at index=%d\nactual: (len=%d)\n%s",
				err, i, len(errs), formatErrorSlice("  - ", errs))
			return
		}
	}
}

func formatErrorSlice(prefix string, errs Errors) string {
	var sb strings.Builder
	for i, err := range errs {
		fmt.Fprintf(&sb, "%s[i=%d] %s\n", prefix, i, err)
	}
	return sb.String()
}
