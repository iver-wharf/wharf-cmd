package tarutil

import (
	"archive/tar"
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDir(t *testing.T) {
	var buf bytes.Buffer
	err := Dir(&buf, "../../test/tarutil/dirtest")
	require.NoError(t, err)

	tr := tar.NewReader(&buf)
	var gotFilenames []string
	for {
		head, err := tr.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		bytes, err := io.ReadAll(tr)
		require.NoError(t, err)
		assert.Equal(t, head.Size, int64(len(bytes)), head.Name)
		gotFilenames = append(gotFilenames, head.Name)
	}
	wantFilenames := []string{
		"bar/",
		"bar/moo",
		"somedir/",
		"somedir/.hidden",
		"somefile.txt",
	}
	assert.ElementsMatch(t, wantFilenames, gotFilenames)
}
