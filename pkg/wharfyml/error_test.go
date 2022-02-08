package wharfyml

import (
	"errors"
	"testing"
)

func requireContainsErr(t *testing.T, errs errorSlice, err error) {
	for _, e := range errs {
		if errors.Is(e, err) {
			return
		}
	}
	t.Fatalf("\nexpected contains error: %q\nactual: (len=%d) %v",
		err, len(errs), errs)
}

func requireNotContainsErr(t *testing.T, errs errorSlice, err error) {
	for i, e := range errs {
		if errors.Is(e, err) {
			t.Fatalf("\nexpected not to contain error: %q\nfound at index=%d\nactual: (len=%d) %v",
				err, i, len(errs), errs)
			return
		}
	}
}
