package ignorer

import (
	"path/filepath"
	"testing"
)

func TestFileIncluder(t *testing.T) {
	i := NewFileIncluder([]string{
		"foo/bar/moo.txt",
		"lorem/ipsum.png",
	})

	t.Run("exact matches", func(t *testing.T) {
		assertIgnore(t, i, false, "foo/bar/moo.txt")
		assertIgnore(t, i, false, "lorem/ipsum.png")
	})

	t.Run("parent directories", func(t *testing.T) {
		assertIgnore(t, i, false, "foo/bar")
		assertIgnore(t, i, false, "foo")
		assertIgnore(t, i, false, "lorem")
	})

	t.Run("ignored files", func(t *testing.T) {
		assertIgnore(t, i, true, "foo/bar/faz.txt")
		assertIgnore(t, i, true, "foo/moo/moo.txt")
		assertIgnore(t, i, true, "lorem/dolor.jpg")
	})

	t.Run("convoluted paths", func(t *testing.T) {
		assertIgnore(t, i, false, "foo/../lorem/ipsum.png")
	})
}

func assertIgnore(t *testing.T, i Ignorer, want bool, path string) {
	path = filepath.Clean(path)
	got := i.Ignore("", path)
	if got != want {
		t.Errorf("want Ignore(%q) = %t, got %t", path, want, got)
	}
}
