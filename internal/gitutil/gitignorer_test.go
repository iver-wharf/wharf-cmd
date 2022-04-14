package gitutil

import (
	"path/filepath"
	"testing"

	"github.com/denormal/go-gitignore"
	"github.com/stretchr/testify/assert"
)

type mockMatcher struct {
	path string
}

func (m *mockMatcher) Match(path string) gitignore.Match {
	// Using filepath.ToSlash so running tests on Windows still works
	m.path = filepath.ToSlash(path)
	return nil
}

func TestGitIgnorer(t *testing.T) {
	mock := &mockMatcher{}
	i := gitIgnorer{
		currentDir: "/home/me/repo/subdir",
		matcher:    mock,
	}

	i.Ignore("foo/bar.txt")
	assert.Equal(t, "/home/me/repo/subdir/foo/bar.txt", mock.path)
}
