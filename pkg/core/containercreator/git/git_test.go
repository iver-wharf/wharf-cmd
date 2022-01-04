package git

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetCommands(t *testing.T) {
	gitContainerCreator := containerCreator{
		containerName: "",
		imageName:     "",
		envVars:       nil,
		volumes:       nil,
		iverCertPath:  "/iverCerts",
	}

	expected := []string{"cp /iverCerts/* /etc/ssl/certs/; " +
		"git config --global http.sslVerify false; " +
		"git clone -b $GIT_SYNC_BRANCH -- $GIT_REPO_URL /gitRepo"}

	commands := gitContainerCreator.GetCommands()
	require.Equal(t, expected, commands)
}
