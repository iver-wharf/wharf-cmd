package errtestutil

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/iver-wharf/wharf-cmd/internal/errutil"
)

// RequireContainsErr fails the test if any error in the slice Is the given
// error.
func RequireContainsErr(t *testing.T, errs errutil.Slice, err error) {
	for _, e := range errs {
		if errors.Is(e, err) {
			return
		}
	}
	t.Fatalf("\nexpected contains error: %q\nactual: (len=%d)\n%s",
		err, len(errs), formatSlice("  - ", errs))
}

// RequireNotContainsErr fails the test if no error in the slice Is the given
// error.
func RequireNotContainsErr(t *testing.T, errs errutil.Slice, err error) {
	for i, e := range errs {
		if errors.Is(e, err) {
			t.Fatalf("\nexpected not to contain error: %q\nfound at index=%d\nactual: (len=%d)\n%s",
				err, i, len(errs), formatSlice("  - ", errs))
			return
		}
	}
}

// RequireNoErr fails the test if the error slice is not empty.
func RequireNoErr(t *testing.T, errs errutil.Slice) {
	if len(errs) == 0 {
		return
	}
	t.Fatalf("\nexpected no errors\nactual: (len=%d)\n%s",
		len(errs), formatSlice("  - ", errs))
}

func formatSlice(prefix string, errs errutil.Slice) string {
	var sb strings.Builder
	for i, err := range errs {
		var posErr errutil.Pos
		if errors.As(err, &posErr) {
			fmt.Fprintf(&sb, "%s[i=%d, at %d:%d] %s\n", prefix, i, posErr.Line, posErr.Column, err)
		} else {
			fmt.Fprintf(&sb, "%s[i=%d] %s\n", prefix, i, err)
		}
	}
	return sb.String()
}
