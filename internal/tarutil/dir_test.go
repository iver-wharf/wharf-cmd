package tarutil

import (
	"archive/tar"
	"bytes"
	"io"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDir(t *testing.T) {
	var buf bytes.Buffer
	err := Dir(&buf, "../../test/tarutil/dirtest")
	require.NoError(t, err)

	gotFilenames := readFilenamesFromTar(t, &buf)
	wantFilenames := []string{
		"bar/",
		"bar/moo",
		"somedir/",
		"somedir/.hidden",
		"somefile.txt",
	}
	assert.ElementsMatch(t, wantFilenames, gotFilenames)
}

type antiBarIgnorer struct{}

func (i antiBarIgnorer) Ignore(path string) bool {
	return filepath.Base(path) == "bar"
}

func TestDirIgnore(t *testing.T) {
	var buf bytes.Buffer
	err := DirIgnore(&buf, "../../test/tarutil/dirtest", antiBarIgnorer{})
	require.NoError(t, err)

	gotFilenames := readFilenamesFromTar(t, &buf)
	wantFilenames := []string{
		// "bar/",
		// "bar/moo",
		"somedir/",
		"somedir/.hidden",
		"somefile.txt",
	}
	assert.ElementsMatch(t, wantFilenames, gotFilenames)
}

func readFilenamesFromTar(t *testing.T, reader io.Reader) []string {
	tr := tar.NewReader(reader)
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
	return gotFilenames
}
