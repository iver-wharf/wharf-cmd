package wharfyml

import (
	"errors"
	"fmt"
	"sort"

	"github.com/goccy/go-yaml/ast"
)

// Pos represents a position. Used to declare where an object was defined in
// the .wharf-ci.yml file. The first line and column starts at 1.
// The zero value is used to represent an undefined position.
type Pos struct {
	Line   int
	Column int
}

// String implements fmt.Stringer
func (p Pos) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}

// IsZero returns true if this type is its zero value.
func (p Pos) IsZero() bool {
	return p == Pos{}
}

func newPosNode(node ast.Node) Pos {
	return Pos{
		Line:   node.GetToken().Position.Line,
		Column: node.GetToken().Position.Column,
	}
}

func wrapPosError(err error, pos Pos) error {
	return PosError{
		Err: err,
		Pos: pos,
	}
}

func wrapPosErrorNode(err error, node ast.Node) error {
	return wrapPosError(err, newPosNode(node))
}

// PosError is an error type that holds metadata about where the error
// occurred (line and column).
type PosError struct {
	Err error
	Pos Pos
}

// Error implements the error interface.
func (err PosError) Error() string {
	if err.Err == nil {
		return ""
	}
	return err.Err.Error()
}

// Is implements the interface to support errors.Is.
func (err PosError) Is(target error) bool {
	return errors.Is(err.Err, target)
}

// Unwrap implements the interface to support errors.Unwrap.
func (err PosError) Unwrap() error {
	return err.Err
}

func posFromError(err error) Pos {
	var posErr PosError
	if !errors.As(err, &posErr) {
		return Pos{}
	}
	return posErr.Pos
}

func sortErrorsByPosition(errs Errors) {
	if len(errs) == 0 {
		return
	}
	sort.Slice(errs, func(i, j int) bool {
		a := posFromError(errs[i])
		b := posFromError(errs[j])
		if a.Line == b.Line {
			return a.Column < b.Column
		}
		return a.Line < b.Line
	})
}
