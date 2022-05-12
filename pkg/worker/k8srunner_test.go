package worker

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIllegalParentDirAccess(t *testing.T) {
	testCases := []struct {
		name   string
		base   string
		path   string
		wantOK bool
	}{
		{
			name:   "rel path not outside current dir - ok",
			path:   filepath.Join(".", "my_path", "to", "file"),
			wantOK: false,
		},
		{
			name:   "rel path not outside current dir - ok",
			path:   filepath.Join(".", "my_path", "..", "my_path_2", "to", "file"),
			wantOK: false,
		},
		{
			name:   "parent dir access - not ok",
			path:   filepath.Join("..", "my_path", "to", "file"),
			wantOK: true,
		},
		{
			name:   "parent dir access, tricky - not ok",
			path:   filepath.Join(".", "my_path", "..", "to", "..", "file", "..", "..", "or", "is", "it"),
			wantOK: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ok := isIllegalParentDirAccess(tc.path)
			assert.Equal(t, tc.wantOK, ok)
		})
	}
}
