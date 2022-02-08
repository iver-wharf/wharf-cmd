package wharfyml

import (
	"errors"
	"fmt"

	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/token"
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

func newParseError(err error, pos *token.Position) error {
	return ParseError{
		Inner:    err,
		Position: pos,
	}
}

func newParseErrorNode(err error, node ast.Node) error {
	return newParseError(err, node.GetToken().Position)
}

type ParseError struct {
	Inner    error
	Position *token.Position
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

func wrapErrorInKeyed(key string, err error) error {
	var keyed keyedError
	if !errors.As(err, &keyed) {
		return keyedError{
			key:   key,
			inner: err,
		}
	}
	return keyedError{
		key:   fmt.Sprintf("%s/%s", key, keyed.key),
		inner: keyed.inner,
	}
}

func wrapErrorSliceInKeyed(key string, errs errorSlice) errorSlice {
	result := make(errorSlice, len(errs))
	for i, err := range errs {
		result[i] = wrapErrorInKeyed(key, err)
	}
	return result
}

type keyedError struct {
	key   string
	inner error
}

func (err keyedError) Error() string {
	if err.inner == nil {
		return err.key
	}
	return fmt.Sprintf("%s: %s", err.key, err.inner)
}

func (err keyedError) Is(target error) bool {
	return errors.Is(err.inner, target)
}

func (err keyedError) Unwrap() error {
	return err.inner
}
