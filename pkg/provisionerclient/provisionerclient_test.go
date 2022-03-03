package provisionerclient

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildURL(t *testing.T) {
	var tests = []struct {
		name  string
		base  string
		paths []string
		want  string
	}{
		{
			name:  "joins",
			base:  "http://localhost",
			paths: []string{"api", "worker"},
			want:  "http://localhost/api/worker",
		},
		{
			name:  "empty segment gives two slashes",
			base:  "http://localhost",
			paths: []string{"api", "", "worker"},
			want:  "http://localhost/api//worker",
		},
		{
			name:  "trims trailing slash",
			base:  "http://localhost/",
			paths: []string{"api", "worker"},
			want:  "http://localhost/api/worker",
		},
		{
			name:  "keeps multiple slashes",
			base:  "http://localhost///",
			paths: []string{"api", "worker"},
			want:  "http://localhost///api/worker",
		},
		{
			name:  "uses base path",
			base:  "http://localhost/api",
			paths: []string{"worker"},
			want:  "http://localhost/api/worker",
		},
		{
			name:  "escapes slashes",
			base:  "http://localhost",
			paths: []string{"api", "worker", "some/id/with/slashes"},
			want:  "http://localhost/api/worker/some%2Fid%2Fwith%2Fslashes",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			u, err := buildURL(tc.base, tc.paths...)
			require.NoError(t, err)
			assert.Equal(t, tc.want, u.String())
		})
	}
}
