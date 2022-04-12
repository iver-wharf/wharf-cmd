package varsub

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testSource = SourceMap{
	"var1": "value1",
	"var2": "value2",
}

func TestReader(t *testing.T) {
	stringReader := strings.NewReader("foo ${var1}\nbar ${var2}")
	r := NewReader(testSource, stringReader)
	b, err := io.ReadAll(r)
	require.NoError(t, err)

	want := "foo value1\nbar value2\n"
	got := string(b)
	assert.Equal(t, want, got)
}

func TestReader_smallBuffer(t *testing.T) {
	stringReader := strings.NewReader("foo ${var1}\nbar ${var2}")
	r := NewReader(testSource, stringReader)

	buf := make([]byte, 7)

	var wantReads = []struct {
		n   int
		str string
	}{
		{7, "foo val"},
		{4, "ue1\n"},
		{7, "bar val"},
		{4, "ue2\n"},
	}

	for i, want := range wantReads {
		n, err := r.Read(buf)
		require.NoErrorf(t, err, "read %d", i+1)
		assert.Equalf(t, want.n, n, "read %d", i+1)
		assert.Equalf(t, want.str, string(buf[:n]), "read %d", i+1)
	}

	n, err := r.Read(buf)
	assert.ErrorIs(t, err, io.EOF, "last read")
	assert.Equal(t, 0, n, "last read")
}
