package wharfyml

import (
	"errors"
	"fmt"

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

func newParseError(err error, line, column int) error {
	return ParseError{
		Inner:  err,
		Line:   line,
		Column: column,
	}
}

func newParseErrorNode(err error, node ast.Node) error {
	pos := node.GetToken().Position
	return newParseError(err, pos.Line, pos.Column)
}

type ParseError struct {
	Inner  error
	Line   int
	Column int
}

func (err ParseError) Error() string {
	if err.Inner == nil {
		return ""
	}
	return err.Inner.Error()
}

func (err ParseError) Is(target error) bool {
	return errors.Is(err.Inner, target)
}

func (err ParseError) Unwrap() error {
	return err.Inner
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
