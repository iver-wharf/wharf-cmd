package wharfyml

import (
	"errors"
	"fmt"
)

// Errors is a slice of errors.
type Errors []error

func (s *Errors) add(errs ...error) {
	*s = append(*s, errs...)
}

func (s *Errors) addNonNils(errs ...error) {
	for _, err := range errs {
		if err == nil {
			continue
		}
		*s = append(*s, err)
	}
}

func wrapPathError(path string, err error) error {
	var keyed pathError
	if !errors.As(err, &keyed) {
		return pathError{
			path:  path,
			inner: err,
		}
	}
	return pathError{
		path:  fmt.Sprintf("%s/%s", path, keyed.path),
		inner: keyed.inner,
	}
}

func wrapPathErrorSlice(path string, errs Errors) Errors {
	result := make(Errors, len(errs))
	for i, err := range errs {
		result[i] = wrapPathError(path, err)
	}
	return result
}

type pathError struct {
	path  string
	inner error
}

// Error implements the error interface.
func (err pathError) Error() string {
	if err.inner == nil {
		return err.path
	}
	return fmt.Sprintf("%s: %s", err.path, err.inner)
}

// Is implements the interface to support errors.Is.
func (err pathError) Is(target error) bool {
	return errors.Is(err.inner, target)
}

// Unwrap implements the interface to support errors.Unwrap.
func (err pathError) Unwrap() error {
	return err.inner
}
