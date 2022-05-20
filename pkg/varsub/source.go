package varsub

import (
	"fmt"

	"github.com/iver-wharf/wharf-cmd/internal/util"
)

// Source is a variable substitution source.
type Source interface {
	// Lookup tries to look up a value based on name and returns that value as
	// well as true on success, or false if the variable was not found.
	Lookup(name string) (Var, bool)

	// ListVars will return a slice of all variables that this varsub Source
	// provides.
	ListVars() []Var
}

// SourceVar is a single Var that also acts as a Source.
type SourceVar Var

// Lookup compares the name and returns the variables value as well as true on
// success, or false if the name does not match.
func (s SourceVar) Lookup(name string) (Var, bool) {
	if s.Key != name {
		return Var{}, false
	}
	return Var(s), true
}

// ListVars will return a slice of only the one variable that this varsub Source
// provides.
func (s SourceVar) ListVars() []Var {
	return []Var{Var(s)}
}

// Var is a single varsub variable, with it's Key (name), Value, and optionally
// also a Source that declares where this variable comes from.
type Var struct {
	Key         string
	Value       any
	SourceLabel string
}

// String implements the fmt.Stringer interface.
func (v Var) String() string {
	return util.Stringify(v.Value)
}

// GoString implements the fmt.GoStringer interface.
func (v Var) GoString() string {
	return fmt.Sprintf("{%q:%[2]T(%#[2]v)}", v.Key, v.Value)
}

// SourceSlice is a slice of variable substution sources that act as a source
// itself by returning the first successful lookup.
type SourceSlice []Source

// Lookup tries to look up a value based on name and returns that value as
// well as true on success, or false if the variable was not found.
func (s SourceSlice) Lookup(name string) (Var, bool) {
	for _, inner := range s {
		val, ok := inner.Lookup(name)
		if ok {
			return val, true
		}
	}
	return Var{}, false
}

// ListVars will return a slice of all variables that this varsub Source
// provides.
func (s SourceSlice) ListVars() []Var {
	var vars []Var
	for _, inner := range s {
		vars = append(vars, inner.ListVars()...)
	}
	return vars
}

// ensure it conforms to interface
var _ Source = SourceSlice{}

// Val is a slimmed down varsub.Var, without the Key, as the SourceMap will
// populate the Key field automatically based on the map keys.
type Val struct {
	Value  any
	Source string
}

// String implements the fmt.Stringer interface.
func (v Val) String() string {
	return util.Stringify(v.Value)
}

// SourceMap is a variable substitution source based on a map where it uses the
// underlying map as the variable source.
type SourceMap map[string]Val

// Lookup tries to look up a value based on name and returns that value as
// well as true on success, or false if the variable was not found.
func (s SourceMap) Lookup(name string) (Var, bool) {
	v, ok := s[name]
	return Var{
		Key:         name,
		Value:       v.Value,
		SourceLabel: v.Source,
	}, ok
}

// ensure it conforms to interface
var _ Source = SourceMap{}

// ListVars will return a slice of all variables that this varsub Source
// provides.
func (s SourceMap) ListVars() []Var {
	var vars []Var
	for k, v := range s {
		vars = append(vars, Var{
			Key:         k,
			Value:       v.Value,
			SourceLabel: v.Source,
		})
	}
	return vars
}
