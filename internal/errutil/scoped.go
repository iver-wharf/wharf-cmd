package errutil

import (
	"errors"
	"strings"
)

const ScopeDelimiter = "/"

// Scope creates a new scoped error. The full scope path can later be retrieved
// from the first scope found via errors.As.
func Scope(err error, paths ...string) error {
	var keyed Scoped
	scope := strings.Join(paths, ScopeDelimiter)
	if !errors.As(err, &keyed) {
		return Scoped{
			scope: scope,
			next:  nil,
			inner: err,
		}
	}
	return Scoped{
		scope: scope,
		next:  &keyed,
		inner: keyed.inner,
	}
}

// ScopeSlice adds a paths to all the errors in the slice.
func ScopeSlice(errs Slice, paths ...string) Slice {
	result := make(Slice, len(errs))
	for i, err := range errs {
		result[i] = Scope(err, paths...)
	}
	return result
}

// AsScope returns the error's scope, or empty string if the error isn't scoped.
func AsScope(err error) string {
	var scoped Scoped
	if errors.As(err, &scoped) {
		return scoped.Scope()
	}
	return ""
}

// Scoped is an error that is scoped. Each scope adds a substring to the
// scope, delimited by a slash.
type Scoped struct {
	scope string
	next  *Scoped
	inner error
}

// Scope returns the joined scope of this error and all inner errors.
func (err Scoped) Scope() string {
	var sb strings.Builder
	scope := &err
	for scope != nil {
		if sb.Len() > 0 {
			sb.WriteString(ScopeDelimiter)
		}
		sb.WriteString(scope.scope)
		scope = scope.next
	}
	return sb.String()
}

// Error implements the error interface.
func (err Scoped) Error() string {
	if err.inner == nil {
		return ""
	}
	return err.inner.Error()
}

// Is implements the interface to support errors.Is.
func (err Scoped) Is(target error) bool {
	return errors.Is(err.inner, target)
}

// Unwrap implements the interface to support errors.Unwrap.
func (err Scoped) Unwrap() error {
	return err.inner
}
