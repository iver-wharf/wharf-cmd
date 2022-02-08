package wharfyml

import (
	"errors"
	"fmt"
	"sort"

	"github.com/goccy/go-yaml/ast"
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

func newPositionedError(err error, line, column int) error {
	return PositionedError{
		Inner:  err,
		Line:   line,
		Column: column,
	}
}

func newPositionedErrorNode(err error, node ast.Node) error {
	pos := node.GetToken().Position
	return newPositionedError(err, pos.Line, pos.Column)
}

// PositionedError is an error type that holds metadata about where the error
// occurred (line and column).
type PositionedError struct {
	Inner  error
	Line   int
	Column int
}

// Error implements the error interface.
func (err PositionedError) Error() string {
	if err.Inner == nil {
		return ""
	}
	return err.Inner.Error()
}

// Is implements the interface to support errors.Is.
func (err PositionedError) Is(target error) bool {
	return errors.Is(err.Inner, target)
}

// Unwrap implements the interface to support errors.Unwrap.
func (err PositionedError) Unwrap() error {
	return err.Inner
}

func positionedErrorLineColumn(err error) (int, int) {
	var parseErr PositionedError
	if !errors.As(err, &parseErr) {
		return 0, 0
	}
	return parseErr.Line, parseErr.Column
}

func sortErrorsByPosition(errs Errors) {
	if len(errs) == 0 {
		return
	}
	sort.Slice(errs, func(i, j int) bool {
		aLine, aCol := positionedErrorLineColumn(errs[i])
		bLine, bCol := positionedErrorLineColumn(errs[j])
		if aLine == bLine {
			return aCol < bCol
		}
		return aLine < bLine
	})
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
