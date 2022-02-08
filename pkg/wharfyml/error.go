package wharfyml

import (
	"errors"

	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/token"
)

type errorSlice []error

func (s *errorSlice) add(errs ...error) {
	*s = append(*s, errs...)
}

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
	if err.Position == nil {
		return err.Inner.Error()
	}
	return err.Inner.Error()
}

func (err ParseError) Is(target error) bool {
	return errors.Is(err.Inner, target)
}

func (err ParseError) Unwrap() error {
	return err.Inner
}
