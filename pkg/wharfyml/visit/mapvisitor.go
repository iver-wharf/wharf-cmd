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

func NewMapVisitor(parent *yaml.Node, nodes map[string]*yaml.Node, source varsub.Source) MapVisitor {
	return MapVisitor{
		parent:    parent,
		nodes:     nodes,
		positions: make(map[string]Pos),
		source:    source,
	}
}

type MapVisitor struct {
	parent    *yaml.Node
	nodes     map[string]*yaml.Node
	positions map[string]Pos
	source    varsub.Source
}

func (p MapVisitor) ParentPos() Pos {
	return NewPosNode(p.parent)
}

func (p MapVisitor) ReadNodesPos() map[string]Pos {
	return p.positions
}

func (p MapVisitor) HasNode(key string) bool {
	_, ok := p.nodes[key]
	return ok
}

func (p MapVisitor) VisitNumber(key string, target *float64) error {
	return visitNode(p, key, target, Float64)
}

func (p MapVisitor) RequireNumberFromVarSub(varLookup string, target *float64) error {
	return requireFromVarSub(p, varLookup, target, p.VisitNumber)
}

func (p MapVisitor) LookupNumberFromVarSub(varLookup string, target *float64) error {
	return lookupFromVarSub(p, varLookup, target, p.VisitNumber)
}

func (p MapVisitor) VisitInt(key string, target *int) error {
	return visitNode(p, key, target, Int)
}

func (p MapVisitor) RequireIntFromVarSub(varLookup string, target *int) error {
	return requireFromVarSub(p, varLookup, target, p.VisitInt)
}

func (p MapVisitor) LookupIntFromVarSub(varLookup string, target *int) error {
	return lookupFromVarSub(p, varLookup, target, p.VisitInt)
}

func (p MapVisitor) VisitString(key string, target *string) error {
	return visitNode(p, key, target, String)
}

func (p MapVisitor) VisitStringSlice(key string, target *[]string) errutil.Slice {
	node, ok := p.nodes[key]
	if !ok {
		return nil
	}
	p.positions[key] = NewPosNode(node)
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

func (p MapVisitor) VisitStringStringMap(key string, target *map[string]string) errutil.Slice {
	node, ok := p.nodes[key]
	if !ok {
		return nil
	}
	p.positions[key] = NewPosNode(node)
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

func (p MapVisitor) VisitStringWithVarSub(nodeKey, varLookup string, target *string) error {
	err := p.loadFromVarSubIfUnset(nodeKey, varLookup)
	if err != nil {
		return err
	}
	return p.VisitString(nodeKey, target)
}

func (p MapVisitor) RequireStringFromVarSub(varLookup string, target *string) error {
	return requireFromVarSub(p, varLookup, target, p.VisitString)
}

func (p MapVisitor) LookupStringFromVarSub(varLookup string, target *string) error {
	return lookupFromVarSub(p, varLookup, target, p.VisitString)
}

func (p MapVisitor) VisitBool(key string, target *bool) error {
	node, ok := p.nodes[key]
	if !ok {
		return nil
	}
	p.positions[key] = NewPosNode(node)
	b, err := Bool(node)
	if err != nil {
		return errutil.Scope(err, key)
	}
	*target = b
	return nil
}

func (p MapVisitor) LookupBoolFromVarSub(varLookup string, target *bool) error {
	return lookupFromVarSub(p, varLookup, target, p.VisitBool)
}

func (p MapVisitor) RequireBoolFromVarSub(varLookup string, target *bool) error {
	return requireFromVarSub(p, varLookup, target, p.VisitBool)
}

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
	return WrapPosErrorNode(inner, p.parent)
}

func visitNode[T any](p MapVisitor, key string, target *T, f func(*yaml.Node) (T, error)) error {
	node, ok := p.nodes[key]
	if !ok {
		return nil
	}
	p.positions[key] = NewPosNode(node)
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
	node, err := p.lookupFromVarSub(varLookup)
	if err != nil {
		return err
	}
	p.nodes[nodeKey] = node
	return nil
}

func (p MapVisitor) lookupFromVarSub(varLookup string) (*yaml.Node, error) {
	varValue, ok := safeLookupVar(p.source, varLookup)
	if !ok {
		err := fmt.Errorf("%w: need %s", ErrMissingBuiltinVar, varLookup)
		return nil, WrapPosErrorNode(err, p.parent)
	}
	newNode, err := NewNodeWithValue(p.parent, varValue.Value)
	if err != nil {
		err := fmt.Errorf("read %s: %w", varLookup, err)
		return nil, WrapPosErrorNode(err, p.parent)
	}
	return newNode, nil
}

func requireFromVarSub[T any](p MapVisitor, varLookup string, target *T, f func(string, *T) error) error {
	node, err := p.lookupFromVarSub(varLookup)
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
