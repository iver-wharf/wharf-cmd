package wharfyml

import (
	"errors"
	"fmt"
	"sort"

	"github.com/goccy/go-yaml/ast"
)

type errorSlice []error

func (s *errorSlice) add(errs ...error) {
	*s = append(*s, errs...)
}

func (s *errorSlice) addNonNils(errs ...error) {
	for _, err := range errs {
		if err == nil {
			continue
		}
		*s = append(*s, err)
	}
}

var fmtErrorfPlaceholder = errors.New("placeholder")

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

type PositionedError struct {
	Inner  error
	Line   int
	Column int
}

func (err PositionedError) Error() string {
	if err.Inner == nil {
		return ""
	}
	return err.Inner.Error()
}

func (err PositionedError) Is(target error) bool {
	return errors.Is(err.Inner, target)
}

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

func sortErrorsByPosition(errs errorSlice) {
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

func wrapPathErrorSlice(path string, errs errorSlice) errorSlice {
	result := make(errorSlice, len(errs))
	for i, err := range errs {
		result[i] = wrapPathError(path, err)
	}
	return result
}

type pathError struct {
	path  string
	inner error
}

func (err pathError) Error() string {
	if err.inner == nil {
		return err.path
	}
	return fmt.Sprintf("%s: %s", err.path, err.inner)
}

func (err pathError) Is(target error) bool {
	return errors.Is(err.inner, target)
}

func (err pathError) Unwrap() error {
	return err.inner
}
