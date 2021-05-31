package wharfyml

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/iver-wharf/wharf-cmd/pkg/core/containercreator/git"
)

func TestGetGitContainer(t *testing.T) {
	repoURL := "https://gitlab.io/testRepoGroup/testRepoName.git"
	gitParams := git.NewGitPropertiesMap(repoURL, "master", "superDUPERtoken")

	gitContainer := GetDefaultContainerSpec(git.NewContainerCreator(git.GetImage(git.DefaultVersion), gitParams, ConfigVolumeMountPath))
	assert.Equal(t, git.ContainerName, gitContainer.Name)
	assert.Equal(t, "alpine/git:v2.30.1", gitContainer.Image)
	assert.Equal(t, 2, len(gitContainer.Env))

	for _, env := range gitContainer.Env {
		assert.Nil(t, env.ValueFrom)
		if env.Name == "GIT_REPO_URL" {
			assert.Equal(t, "https://oauth2:superDUPERtoken@gitlab.io/testRepoGroup/testRepoName.git", env.Value)
		} else if env.Name == "GIT_SYNC_BRANCH" {
			assert.Equal(t, "master", env.Value)
		} else {
			t.Error(env.Name, env.Value)
		}
	}
	assert.Equal(t, 2, len(gitContainer.Args))
	assert.Equal(t, 2, len(gitContainer.VolumeMounts))
	assert.Equal(t, "cp /iverCerts/* /etc/ssl/certs/;"+
		" git config --global http.sslVerify false;"+
		" git clone -b $GIT_SYNC_BRANCH -- $GIT_REPO_URL /gitRepo",
		gitContainer.Args[1])
}
