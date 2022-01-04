package git

import (
	"fmt"
	"strings"
)

const RepoDest = "/gitRepo"

type EnvVar int

const (
	RepoURL = EnvVar(iota + 1)
	SyncBranch
)

func (e EnvVar) String() string {
	switch e {
	case RepoURL:
		return "GIT_REPO_URL"
	case SyncBranch:
		return "GIT_SYNC_BRANCH"
	}
	return ""
}

func NewGitPropertiesMap(URL string, branch string, token string) map[EnvVar]string {
	oauthPrefix := fmt.Sprintf("https://oauth2:%s@", token)

	m := make(map[EnvVar]string)

	m[RepoURL] = strings.Replace(URL, "https://", oauthPrefix, 1)
	m[SyncBranch] = branch
	return m
}
