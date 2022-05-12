package yamltesting

import (
	"fmt"
	"testing"

	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml/visit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func NewKeyedNode(t *testing.T, content string) (visit.StringNode, *yaml.Node) {
	t.Helper()
	node := NewNode(t, content)
	require.Equal(t, yaml.MappingNode, node.Kind, "keyed node")
	require.Len(t, node.Content, 2, "keyed node")
	require.Equal(t, yaml.ScalarNode, node.Content[0].Kind, "key node kind in keyed node")
	require.Equal(t, visit.ShortTagString, node.Content[0].ShortTag(), "key node tag in keyed node")
	return visit.StringNode{Node: node.Content[0], Value: node.Content[0].Value}, node.Content[1]
}

func NewNode(t *testing.T, content string) *yaml.Node {
	t.Helper()
	var doc yaml.Node
	err := yaml.Unmarshal([]byte(content), &doc)
	require.NoError(t, err, "parse node")
	require.Equal(t, yaml.DocumentNode, doc.Kind, "document node")
	require.Len(t, doc.Content, 1, "document node count")
	return doc.Content[0]
}

func AssertVarSubNode(t *testing.T, want any, actual visit.VarSubNode, messageAndArgs ...any) {
	t.Helper()
	var value any
	var err error
	switch actual.Node.ShortTag() {
	case visit.ShortTagBool:
		value, err = visit.Bool(actual.Node)
	case visit.ShortTagInt:
		value, err = visit.Int(actual.Node)
	case visit.ShortTagFloat:
		value, err = visit.Float64(actual.Node)
	case visit.ShortTagString:
		value, err = visit.String(actual.Node)
	default:
		assert.Fail(t, fmt.Sprintf("expected string, boolean, or number, but found %s",
			visit.PrettyNodeTypeName(actual.Node)), messageAndArgs...)
		return
	}
	if err != nil {
		assert.Fail(t, err.Error(), messageAndArgs...)
		return
	}
	assert.Equal(t, want, value, messageAndArgs...)
}
