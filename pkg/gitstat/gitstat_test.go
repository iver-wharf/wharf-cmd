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

func TestEstimateRepoGroupAndName(t *testing.T) {
	tests := []struct {
		name      string
		origin    Remote
		wantGroup string
		wantName  string
	}{
		{
			name:      "empty",
			origin:    Remote{},
			wantGroup: "",
			wantName:  "",
		},
		{
			name:      "github/ssh-path",
			origin:    newTestRemote("git@github.com:iver-wharf/wharf-cmd.git"),
			wantGroup: "iver-wharf",
			wantName:  "wharf-cmd",
		},
		{
			name:      "github/ssh-url",
			origin:    newTestRemote("ssh://git@github.com/iver-wharf/wharf-cmd.git"),
			wantGroup: "iver-wharf",
			wantName:  "wharf-cmd",
		},
		{
			name:      "github/http-no-dotgit",
			origin:    newTestRemote("http://git@github.com/iver-wharf/wharf-cmd"),
			wantGroup: "iver-wharf",
			wantName:  "wharf-cmd",
		},
		{
			name:      "dev.azure.com/ssh-path",
			origin:    newTestRemote("git@ssh.dev.azure.com:v3/iver-wharf/wharf/wharf-cmd"),
			wantGroup: "iver-wharf/wharf",
			wantName:  "wharf-cmd",
		},
		{
			name:      "dev.azure.com/https",
			origin:    newTestRemote("https://iver-wharf@dev.azure.com/iver-wharf/wharf/_git/wharf-cmd"),
			wantGroup: "iver-wharf/wharf",
			wantName:  "wharf-cmd",
		},
		{
			name:      "gitlab/ssh-path",
			origin:    newTestRemote("git@gitlab.com:iver-wharf/wharf-subgroup/wharf-cmd.git"),
			wantGroup: "iver-wharf/wharf-subgroup",
			wantName:  "wharf-cmd",
		},
		{
			name:      "gitlab/https-url",
			origin:    newTestRemote("https://gitlab.com/iver-wharf/wharf-subgroup/wharf-cmd.git"),
			wantGroup: "iver-wharf/wharf-subgroup",
			wantName:  "wharf-cmd",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotGroup, gotName := estimateRepoGroupAndName(tc.origin)
			assert.Equal(t, tc.wantGroup, gotGroup, "group name")
			assert.Equal(t, tc.wantName, gotName, "repo name")
		})
	}
}

func newTestRemote(url string) Remote {
	return Remote{
		FetchURL: url,
		PushURL:  url,
	}
}
