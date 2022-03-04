package httptests

import (
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/iver-wharf/wharf-cmd/pkg/workerapi/workerclient"
	"github.com/iver-wharf/wharf-cmd/pkg/workerapi/workerserver"
	"github.com/iver-wharf/wharf-cmd/pkg/workerapi/workerserver/model/response"
	"github.com/stretchr/testify/assert"
)

const (
	serverBindAddress   = "0.0.0.0:8080"
	clientTargetAddress = "127.0.0.1:8080"
	insecureSkipVerify  = true
)

func TestListBuildSteps(t *testing.T) {
	server := launchServer()
	defer server.ForceStop()
	client := newClient(t)

	wantSteps := []response.Step{
		{
			Name:     "step-1",
			StepType: response.StepType{Name: "container"},
		},
		{
			Name:     "step-2",
			StepType: response.StepType{Name: "container"},
		},
	}

	gotSteps, err := client.ListBuildSteps()
	assert.NoError(t, err)
	assert.Equal(t, wantSteps, gotSteps)
}

func TestListArtifacts(t *testing.T) {
	server := launchServer()
	defer server.ForceStop()
	client := newClient(t)

	wantArtifacts := []response.Artifact{
		{
			ArtifactID: validArtifactID1,
			StepID:     1,
			Name:       validArtifactName1,
		},
		{
			ArtifactID: validArtifactID2,
			StepID:     1,
			Name:       validArtifactName2,
		},
	}

	gotArtifacts, err := client.ListArtifacts()
	assert.NoError(t, err)
	assert.Equal(t, wantArtifacts, gotArtifacts)
}

func TestDownloadArtifact(t *testing.T) {
	server := launchServer()
	defer server.GracefulStop()
	client := newClient(t)

	testCases := []struct {
		name       string
		artifactID uint
		wantData   []byte
		wantErr    error
	}{
		{
			name:       "get existing artifact 1 works",
			artifactID: validArtifactID1,
			wantData:   artifactData1,
			wantErr:    nil,
		},
		{
			name:       "get existing artifact 2 works",
			artifactID: validArtifactID2,
			wantData:   artifactData2,
			wantErr:    nil,
		},
		{
			name:       "get non-existing artifact fails",
			artifactID: invalidArtifactID,
			wantData:   nil,
			wantErr:    errArtifactNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ioBody, gotErr := client.DownloadArtifact(tc.artifactID)
			if tc.wantErr != nil {
				assert.EqualError(t, gotErr, fmt.Sprintf("(problem) record not found: %s", tc.wantErr.Error()))
			}
			if tc.wantData == nil {
				assert.Nil(t, ioBody)
			} else {
				gotData, _ := io.ReadAll(ioBody)
				assert.Equal(t, tc.wantData, gotData)
			}
		})
	}
}

func launchServer() workerserver.Server {
	server := workerserver.NewHTTPServer(serverBindAddress, &mockBuildStepLister{}, &mockArtifactLister{}, &mockArtifactDownloader{})
	go func() {
		if err := server.Serve(); err != nil {
			panic(err)
		}
	}()

	for !server.IsRunning() {
		time.Sleep(time.Millisecond)
	}
	return server
}

func newClient(t *testing.T) workerclient.HTTPClient {
	client, err := workerclient.NewClient(clientTargetAddress, insecureSkipVerify)
	assert.NoError(t, err)
	return client
}