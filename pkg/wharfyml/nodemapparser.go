package wharfyml

import (
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"
)

// Errors related to parsing map of nodes.
var (
	ErrMissingRequired = errors.New("missing required field")
)

func newNodeMapParser(parent *yaml.Node, nodes map[string]*yaml.Node) nodeMapParser {
	return nodeMapParser{
		parent:    parent,
		nodes:     nodes,
		positions: make(map[string]Pos),
	}
}

type nodeMapParser struct {
	parent    *yaml.Node
	nodes     map[string]*yaml.Node
	positions map[string]Pos
}

func (p nodeMapParser) parentPos() Pos {
	return newPosNode(p.parent)
}

func (p nodeMapParser) unmarshalNumber(key string, target *float64) error {
	node, ok := p.nodes[key]
	if !ok {
		return nil
	}
	p.positions[key] = newPosNode(node)
	num, err := visitFloat64(node)
	if err != nil {
		return wrapPathError(key, err)
	}
	*target = num
	return nil
}

func (p nodeMapParser) unmarshalString(key string, target *string) error {
	node, ok := p.nodes[key]
	if !ok {
		return nil
	}
	p.positions[key] = newPosNode(node)
	str, err := visitString(node)
	if err != nil {
		return wrapPathError(key, err)
	}
	*target = str
	return nil
}

func (p nodeMapParser) unmarshalStringSlice(key string, target *[]string) Errors {
	node, ok := p.nodes[key]
	if !ok {
		return nil
	}
	p.positions[key] = newPosNode(node)
	seq, err := visitSequence(node)
	if err != nil {
		return Errors{err}
	}
	strs := make([]string, 0, len(seq))
	var errSlice Errors
	for i, n := range seq {
		str, err := visitString(n)
		if err != nil {
			errSlice.add(wrapPathError(fmt.Sprintf("%s[%d]", key, i), err))
			continue
		}
		strs = append(strs, str)
	}
	*target = strs
	return errSlice
}

func (p nodeMapParser) unmarshalStringStringMap(key string, target *map[string]string) Errors {
	node, ok := p.nodes[key]
	if !ok {
		return nil
	}
	p.positions[key] = newPosNode(node)
	var errSlice Errors
	nodes, errs := visitMapSlice(node)
	errSlice.add(wrapPathErrorSlice(key, errs)...)

	strMap := make(map[string]string, len(nodes))
	for _, n := range nodes {
		val, err := visitString(n.value)
		if err != nil {
			errSlice.add(wrapPathError(fmt.Sprintf("%s.%s", key, n.key.value), err))
			continue
		}
		strMap[n.key.value] = val
	}
	*target = strMap
	return errSlice
}

func (p nodeMapParser) unmarshalBool(key string, target *bool) error {
	node, ok := p.nodes[key]
	if !ok {
		return nil
	}
	p.positions[key] = newPosNode(node)
	b, err := visitBool(node)
	if err != nil {
		return wrapPathError(key, err)
	}
	*target = b
	return nil
}

func (p nodeMapParser) validateRequiredString(key string) error {
	node, ok := p.nodes[key]
	if !ok {
		return p.newRequiredError(key)
	}
	isStr := node.Kind == yaml.ScalarNode && node.ShortTag() == shortTagString
	if isStr && node.Value == "" {
		return p.newRequiredError(key)
	}
	return nil
}

func (p nodeMapParser) validateRequiredSlice(key string) error {
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

func (p nodeMapParser) newRequiredError(key string) error {
	inner := fmt.Errorf("%w: %q", ErrMissingRequired, key)
	return wrapPosErrorNode(inner, p.parent)
}

func newInvalidFieldTypeErr(key string, wantType string, node *yaml.Node) error {
	gotType := prettyNodeTypeName(node)
	err := wrapPosErrorNode(fmt.Errorf("%w: expected %s, but found %s",
		ErrInvalidFieldType, wantType, gotType), node)
	return wrapPathError(key, err)
}
