package errutil

import (
	"errors"
	"sort"

	"gopkg.in/yaml.v3"
)

// NewPos wraps an error with a position. If added to an error that already has
// a position, then that inner position is shadowed by this new position,
// rendering it hidden/ignored.
func NewPos(err error, line, column int) error {
	return Pos{
		Err:    err,
		Line:   line,
		Column: column,
	}
}

// NewPosFromNode is a utility function that creates a new Pos error based on the
// position info from a YAML node.
func NewPosFromNode(err error, node *yaml.Node) error {
	return Pos{
		Err:    err,
		Line:   node.Line,
		Column: node.Column,
	}
}

// Pos is a positioned error type that holds metadata about where the error
// occurred (line and column).
type Pos struct {
	Err    error
	Line   int
	Column int
}

// Error implements the error interface.
func (err Pos) Error() string {
	if err.Err == nil {
		return ""
	}
	return err.Err.Error()
}

// Is implements the interface to support errors.Is.
func (err Pos) Is(target error) bool {
	return errors.Is(err.Err, target)
}

// Unwrap implements the interface to support errors.Unwrap.
func (err Pos) Unwrap() error {
	return err.Err
}

// AsPos returns the position of the error, or 0,0 if the error doesn't have a
// position.
//
// Only the first position found is used. I.e. shadowed Pos errors are ignored.
func AsPos(err error) (int, int) {
	var posErr Pos
	if !errors.As(err, &posErr) {
		return 0, 0
	}
	return posErr.Line, posErr.Column
}

// SortByPos sorts a slice of errors by their position. Errors without a
// position are placed first in the list in arbitrary order.
func SortByPos(errs Slice) {
	if len(errs) == 0 {
		return
	}
	sort.Slice(errs, func(i, j int) bool {
		aLine, aColumn := AsPos(errs[i])
		bLine, bColumn := AsPos(errs[j])
		if aLine == bLine {
			return aColumn < bColumn
		}
		return aLine < bLine
	})
}
