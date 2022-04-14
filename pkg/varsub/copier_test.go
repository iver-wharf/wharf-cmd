package varsub

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testSource = SourceMap{
	"var1": "value1",
	"var2": "value2",
}

func TestCopier(t *testing.T) {
	stringReader := strings.NewReader("foo ${var1}\nbar ${var2}")
	c := NewCopier(testSource)
	var buf bytes.Buffer

	err := c.Copy(&buf, stringReader)
	require.NoError(t, err)

	want := "foo value1\nbar value2\n"
	got := buf.String()
	assert.Equal(t, want, got)
}
