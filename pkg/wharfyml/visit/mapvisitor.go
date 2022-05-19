package visit

import (
	"errors"
	"fmt"

	"github.com/iver-wharf/wharf-cmd/internal/errutil"
	"github.com/iver-wharf/wharf-cmd/pkg/varsub"
	"gopkg.in/yaml.v3"
)

// Errors for the MapVisitor
var (
	ErrMissingRequired   = errors.New("missing required field")
	ErrMissingBuiltinVar = errors.New("missing built-in var")
)

// NewMapVisitor creates a new type that can visit nodes from a map of nodes by
// key, allowing simpler parsing of a map node with many different field types.
func NewMapVisitor(parent *yaml.Node, nodes map[string]*yaml.Node, source varsub.Source) MapVisitor {
	return MapVisitor{
		parent:    parent,
		nodes:     nodes,
		positions: make(map[string]Pos),
		source:    source,
	}
}

// MapVisitor is a utility type that can visit nodes from a map of nodes by
// key, allowing simpler parsing of a map node with many different field types.
type MapVisitor struct {
	parent    *yaml.Node
	nodes     map[string]*yaml.Node
	positions map[string]Pos
	source    varsub.Source
}

// ParentPos returns the position of the map node's parent.
func (p MapVisitor) ParentPos() Pos {
	return NewPosFromNode(p.parent)
}

// ReadNodesPos returns a map of the positions of all the nodes that have been
// visited so far.
func (p MapVisitor) ReadNodesPos() map[string]Pos {
	return p.positions
}

// HasNode returns a boolean if the node is defined. A YAML node with the value
// null will still return true.
func (p MapVisitor) HasNode(key string) bool {
	_, ok := p.nodes[key]
	return ok
}

// VisitNumber reads a node by string key and writes the parsed float64 value to
// the pointer. An error is returned on parse error. If the node is not present,
// then nil is returned and the pointer is untouched.
func (p MapVisitor) VisitNumber(key string, target *float64) error {
	return visitNode(p, key, target, Float64)
}

// RequireNumberFromVarSub looks up a value from the predefined varsub.Source
// and writes the value to the pointer. An error is returned if the looked up
// variable is not found or on type errors.
func (p MapVisitor) RequireNumberFromVarSub(varLookup string, target *float64) error {
	return requireFromVarSub(p, varLookup, target, p.VisitNumber)
}

// LookupNumberFromVarSub looks up a value from the predefined varsub.Source
// and writes the value to the pointer. An error is returned on type error.
// If the variable is not present, then nil is returned and the pointer is
// untouched.
func (p MapVisitor) LookupNumberFromVarSub(varLookup string, target *float64) error {
	return lookupFromVarSub(p, varLookup, target, p.VisitNumber)
}

// VisitInt reads a node by string key and writes the parsed int value to
// the pointer. An error is returned on parse error. If the node is not present,
// then nil is returned and the pointer is untouched.
func (p MapVisitor) VisitInt(key string, target *int) error {
	return visitNode(p, key, target, Int)
}

// RequireIntFromVarSub looks up a value from the predefined varsub.Source
// and writes the value to the pointer. An error is returned if the looked up
// variable is not found or on type errors.
func (p MapVisitor) RequireIntFromVarSub(varLookup string, target *int) error {
	return requireFromVarSub(p, varLookup, target, p.VisitInt)
}

// LookupIntFromVarSub looks up a value from the predefined varsub.Source
// and writes the value to the pointer. An error is returned on type error.
// If the variable is not present, then nil is returned and the pointer is
// untouched.
func (p MapVisitor) LookupIntFromVarSub(varLookup string, target *int) error {
	return lookupFromVarSub(p, varLookup, target, p.VisitInt)
}

// VisitUint reads a node by string key and writes the parsed int value to
// the pointer. An error is returned on parse error. If the node is not present,
// then nil is returned and the pointer is untouched.
func (p MapVisitor) VisitUint(key string, target *uint) error {
	return visitNode(p, key, target, Uint)
}

// RequireUintFromVarSub looks up a value from the predefined varsub.Source
// and writes the value to the pointer. An error is returned if the looked up
// variable is not found or on type errors.
func (p MapVisitor) RequireUintFromVarSub(varLookup string, target *uint) error {
	return requireFromVarSub(p, varLookup, target, p.VisitUint)
}

// LookupUintFromVarSub looks up a value from the predefined varsub.Source
// and writes the value to the pointer. An error is returned on type error.
// If the variable is not present, then nil is returned and the pointer is
// untouched.
func (p MapVisitor) LookupUintFromVarSub(varLookup string, target *uint) error {
	return lookupFromVarSub(p, varLookup, target, p.VisitUint)
}

// VisitString reads a node by string key and writes the string value to
// the pointer. An error is returned on type error. If the node is not present,
// then nil is returned and the pointer is untouched.
func (p MapVisitor) VisitString(key string, target *string) error {
	return visitNode(p, key, target, String)
}

// VisitStringSlice reads a node by string key and writes the string values to
// the pointer. A slice of error contains any type errors. If the node is not
// present, then nil is returned and the pointer is untouched.
func (p MapVisitor) VisitStringSlice(key string, target *[]string) errutil.Slice {
	node, ok := p.nodes[key]
	if !ok {
		return nil
	}
	p.positions[key] = NewPosFromNode(node)
	seq, err := Sequence(node)
	if err != nil {
		return errutil.Slice{errutil.Scope(err, key)}
	}
	strs := make([]string, 0, len(seq))
	var errSlice errutil.Slice
	for i, n := range seq {
		str, err := String(n)
		if err != nil {
			errSlice.Add(errutil.Scope(err, fmt.Sprintf("%s[%d]", key, i)))
			continue
		}
		strs = append(strs, str)
	}
	*target = strs
	return errSlice
}

// VisitStringStringMap reads a node by string key and writes the string
// key-value pairs to the pointer. A slice of error contains any type errors. If
// the node is not present, then nil is returned and the pointer is untouched.
func (p MapVisitor) VisitStringStringMap(key string, target *map[string]string) errutil.Slice {
	node, ok := p.nodes[key]
	if !ok {
		return nil
	}
	p.positions[key] = NewPosFromNode(node)
	var errSlice errutil.Slice
	nodes, errs := MapSlice(node)
	errSlice.Add(errutil.ScopeSlice(errs, key)...)

	strMap := make(map[string]string, len(nodes))
	for _, n := range nodes {
		val, err := String(n.Value)
		if err != nil {
			errSlice.Add(errutil.Scope(err, fmt.Sprintf("%s.%s", key, n.Key.Value)))
			continue
		}
		strMap[n.Key.Value] = val
	}
	*target = strMap
	return errSlice
}

// VisitStringWithVarSub will try to find a variable from the predefined
// varsub.Source and use that as a default if the node at the given key is not
// present. An error is returned on type errors.
func (p MapVisitor) VisitStringWithVarSub(nodeKey, varLookup string, target *string) error {
	err := p.loadFromVarSubIfUnset(nodeKey, varLookup)
	if err != nil {
		return err
	}
	return p.VisitString(nodeKey, target)
}

// RequireStringFromVarSub looks up a value from the predefined varsub.Source
// and writes the value to the pointer. An error is returned if the looked up
// variable is not found or on type errors.
func (p MapVisitor) RequireStringFromVarSub(varLookup string, target *string) error {
	return requireFromVarSub(p, varLookup, target, p.VisitString)
}

// LookupStringFromVarSub looks up a value from the predefined varsub.Source
// and writes the value to the pointer. An error is returned on type error.
// If the variable is not present, then nil is returned and the pointer is
// untouched.
func (p MapVisitor) LookupStringFromVarSub(varLookup string, target *string) error {
	return lookupFromVarSub(p, varLookup, target, p.VisitString)
}

// VisitBool reads a node by string key and writes the parsed bool value to
// the pointer. An error is returned on parse error. If the node is not present,
// then nil is returned and the pointer is untouched.
func (p MapVisitor) VisitBool(key string, target *bool) error {
	return visitNode(p, key, target, Bool)
}

// LookupBoolFromVarSub looks up a value from the predefined varsub.Source
// and writes the value to the pointer. An error is returned on type error.
// If the variable is not present, then nil is returned and the pointer is
// untouched.
func (p MapVisitor) LookupBoolFromVarSub(varLookup string, target *bool) error {
	return lookupFromVarSub(p, varLookup, target, p.VisitBool)
}

// RequireBoolFromVarSub looks up a value from the predefined varsub.Source
// and writes the value to the pointer. An error is returned if the looked up
// variable is not found or on type errors.
func (p MapVisitor) RequireBoolFromVarSub(varLookup string, target *bool) error {
	return requireFromVarSub(p, varLookup, target, p.VisitBool)
}

// ValidateRequiredString will return an error if the node at the given key is
// not found or is an empty string.
func (p MapVisitor) ValidateRequiredString(key string) error {
	node, ok := p.nodes[key]
	if !ok {
		return p.newRequiredError(key)
	}
	isStr := node.Kind == yaml.ScalarNode && node.ShortTag() == ShortTagString
	if isStr && node.Value == "" {
		return p.newRequiredError(key)
	}
	return nil
}

// ValidateRequiredSlice will return an error if the node at the given key is
// not found or is an empty slice.
func (p MapVisitor) ValidateRequiredSlice(key string) error {
	node, ok := p.nodes[key]
	if !ok {
		return p.newRequiredError(key)
	}
	isSeq := node.Kind == yaml.SequenceNode
	if isSeq && len(node.Content) == 0 {
		return p.newRequiredError(key)
	}
	return nil
}

func (p MapVisitor) newRequiredError(key string) error {
	inner := fmt.Errorf("%w: %q", ErrMissingRequired, key)
	return errutil.NewPosFromNode(inner, p.parent)
}

func visitNode[T any](p MapVisitor, key string, target *T, f func(*yaml.Node) (T, error)) error {
	node, ok := p.nodes[key]
	if !ok {
		return nil
	}
	p.positions[key] = NewPosFromNode(node)
	val, err := f(node)
	if err != nil {
		return errutil.Scope(err, key)
	}
	*target = val
	return nil
}

func (p MapVisitor) loadFromVarSubIfUnset(nodeKey, varLookup string) error {
	if _, ok := p.nodes[nodeKey]; ok {
		return nil
	}
	node, err := p.requireFromVarSub(varLookup)
	if err != nil {
		return err
	}
	p.nodes[nodeKey] = node
	return nil
}

func (p MapVisitor) requireFromVarSub(varLookup string) (*yaml.Node, error) {
	varValue, ok := safeLookupVar(p.source, varLookup)
	if !ok {
		err := fmt.Errorf("%w: need %s", ErrMissingBuiltinVar, varLookup)
		return nil, errutil.NewPosFromNode(err, p.parent)
	}
	if varValue.Value == "" {
		err := fmt.Errorf("%w: empty %s", ErrMissingBuiltinVar, varLookup)
		return nil, errutil.NewPosFromNode(err, p.parent)
	}
	newNode, err := NewNodeWithValue(p.parent, varValue.Value)
	if err != nil {
		err := fmt.Errorf("read %s: %w", varLookup, err)
		return nil, errutil.NewPosFromNode(err, p.parent)
	}
	return newNode, nil
}

func requireFromVarSub[T any](p MapVisitor, varLookup string, target *T, f func(string, *T) error) error {
	node, err := p.requireFromVarSub(varLookup)
	if err != nil {
		return err
	}
	p.nodes["__tmp"] = node
	err = f("__tmp", target)
	delete(p.nodes, "__tmp")
	return err
}

func lookupFromVarSub[T any](p MapVisitor, varLookup string, target *T, f func(string, *T) error) error {
	err := requireFromVarSub(p, varLookup, target, f)
	if errors.Is(err, ErrMissingBuiltinVar) {
		return nil
	}
	return err
}

// AddErrorFor will add an error to the slice of errors with errutil.Pos and
// errutil.Scoped errors wrapped around it, using the node's position and
// key name respectively.
func (p MapVisitor) AddErrorFor(key string, errSlice *errutil.Slice, err error) {
	node, ok := p.nodes[key]
	if ok {
		err = errutil.NewPosFromNode(err, node)
	} else {
		if p.parent != nil {
			err = errutil.NewPosFromNode(err, p.parent)
		}
	}
	err = errutil.Scope(err, key)
	errSlice.Add(err)
}
