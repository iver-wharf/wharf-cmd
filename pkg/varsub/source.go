package varsub

import "fmt"

// Source is a variable substitution source.
type Source interface {
	// Lookup tries to look up a value based on name and returns that value as
	// well as true on success, or false if the variable was not found.
	Lookup(name string) (Var, bool)

	ListVars() []Var
}

type Var struct {
	Key    string
	Value  any
	Source string
}

func (v Var) String() string {
	return stringify(v.Value)
}

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

func (s SourceSlice) ListVars() []Var {
	var vars []Var
	for _, inner := range s {
		vars = append(vars, inner.ListVars()...)
	}
	return vars
}

// ensure it conforms to interface
var _ Source = SourceSlice{}

type Val struct {
	Value  any
	Source string
}

func (v Val) String() string {
	return stringify(v.Value)
}

// SourceMap is a variable substitution source based on a map where it uses the
// underlying map as the variable source.
type SourceMap map[string]Val

// Lookup tries to look up a value based on name and returns that value as
// well as true on success, or false if the variable was not found.
func (s SourceMap) Lookup(name string) (Var, bool) {
	v, ok := s[name]
	return Var{
		Key:    name,
		Value:  v.Value,
		Source: v.Source,
	}, ok
}

// ensure it conforms to interface
var _ Source = SourceMap{}

func (s SourceMap) ListVars() []Var {
	var vars []Var
	for k, v := range s {
		vars = append(vars, Var{
			Key:    k,
			Value:  v.Value,
			Source: v.Source,
		})
	}
	return vars
}
