package gitstat

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseRemotes(t *testing.T) {
	tests := []struct {
		name  string
		lines []string
		want  map[string]Remote
	}{
		{
			name:  "empty",
			lines: nil,
			want:  map[string]Remote{},
		},
		{
			name: "single push and fetch",
			lines: []string{
				"origin\tgit@github.com:iver-wharf/wharf-cmd.git (fetch)",
				"origin\tgit@github.com:iver-wharf/wharf-cmd.git (push)",
			},
			want: map[string]Remote{
				"origin": {
					FetchURL: "git@github.com:iver-wharf/wharf-cmd.git",
					PushURL:  "git@github.com:iver-wharf/wharf-cmd.git",
				},
			},
		},
		{
			name: "reverse order",
			lines: []string{
				"origin\tgit@github.com:iver-wharf/wharf-cmd.git (push)",
				"origin\tgit@github.com:iver-wharf/wharf-cmd.git (fetch)",
			},
			want: map[string]Remote{
				"origin": {
					FetchURL: "git@github.com:iver-wharf/wharf-cmd.git",
					PushURL:  "git@github.com:iver-wharf/wharf-cmd.git",
				},
			},
		},
		{
			name: "only push",
			lines: []string{
				"origin\tgit@github.com:iver-wharf/wharf-cmd.git (push)",
			},
			want: map[string]Remote{
				"origin": {PushURL: "git@github.com:iver-wharf/wharf-cmd.git"},
			},
		},
		{
			name: "only fetch",
			lines: []string{
				"origin\tgit@github.com:iver-wharf/wharf-cmd.git (fetch)",
			},
			want: map[string]Remote{
				"origin": {FetchURL: "git@github.com:iver-wharf/wharf-cmd.git"},
			},
		},
		{
			name: "multiple",
			lines: []string{
				"origin\tgit@github.com:iver-wharf/wharf-cmd-fork.git (fetch)",
				"origin\tgit@github.com:iver-wharf/wharf-cmd-fork.git (push)",
				"upstream\tgit@github.com:iver-wharf/wharf-cmd.git (fetch)",
				"upstream\tgit@github.com:iver-wharf/wharf-cmd.git (push)",
			},
			want: map[string]Remote{
				"origin": {
					FetchURL: "git@github.com:iver-wharf/wharf-cmd-fork.git",
					PushURL:  "git@github.com:iver-wharf/wharf-cmd-fork.git",
				},
				"upstream": {
					FetchURL: "git@github.com:iver-wharf/wharf-cmd.git",
					PushURL:  "git@github.com:iver-wharf/wharf-cmd.git",
				},
			},
		},
		{
			name: "skips invalid",
			lines: []string{
				"origin\tgit@github.com:iver-wharf/wharf-cmd.git (fetch)",
				"foo bar",
				"origin\tgit@github.com:iver-wharf/wharf-cmd.git (push)",
			},
			want: map[string]Remote{
				"origin": {
					FetchURL: "git@github.com:iver-wharf/wharf-cmd.git",
					PushURL:  "git@github.com:iver-wharf/wharf-cmd.git",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := parseRemotes(tc.lines)
			assert.Equal(t, tc.want, got)
		})
	}
}
