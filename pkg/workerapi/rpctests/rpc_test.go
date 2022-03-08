package rpctests

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/iver-wharf/wharf-cmd/pkg/workerapi/workerclient"
	"github.com/iver-wharf/wharf-cmd/pkg/workerapi/workerserver"
	"github.com/stretchr/testify/assert"
)

const (
	serverBindAddress   = "127.0.0.1:8081"
	clientTargetAddress = "127.0.0.1:8081"
)

func TestStreamStatusEvents(t *testing.T) {
	server := launchServer(t)
	defer server.ForceStop()
	client, err := launchClient()
	assert.NoError(t, err)
	defer client.Close()
	wantStatusEventsCh, err := new(mockStore).SubAllStatusUpdates(100)
	assert.NoError(t, err)
	var wantStatusEvents []*workerclient.StatusEvent
	for statusEvent := range wantStatusEventsCh {
		wantStatusEvents = append(wantStatusEvents, workerserver.ConvertToStreamStatusEventsResponse(statusEvent))
	}

	stream, err := client.StreamStatusEvents(context.Background(), &workerclient.StatusEventsRequest{})
	assert.NoError(t, err)

	var gotStatusEvents []*workerclient.StatusEvent
	for {
		statusEvent, err := stream.Recv()
		if err == io.EOF {
			break
		}
		assert.NoError(t, err)
		gotStatusEvents = append(gotStatusEvents, statusEvent)
	}
	for k, v := range wantStatusEvents {
		assert.Equal(t, v.EventID, gotStatusEvents[k].EventID)
		assert.Equal(t, v.Status, gotStatusEvents[k].Status)
		assert.Equal(t, v.StepID, gotStatusEvents[k].StepID)
	}
}

func TestStreamLogs(t *testing.T) {
	server := launchServer(t)
	defer server.ForceStop()
	client, err := launchClient()
	assert.NoError(t, err)
	defer client.Close()
	wantLogsCh, err := new(mockStore).SubAllLogLines(100)
	assert.NoError(t, err)
	var wantLogs []*workerclient.LogLine
	for line := range wantLogsCh {
		wantLogs = append(wantLogs, workerserver.ConvertToStreamLogsResponse(line))
	}

	stream, err := client.StreamLogs(context.Background(), &workerclient.LogsRequest{})
	assert.NoError(t, err)

	var gotLogs []*workerclient.LogLine
	for {
		line, err := stream.Recv()
		if err == io.EOF {
			break
		}
		assert.NoError(t, err)
		gotLogs = append(gotLogs, line)
	}

	for k, v := range wantLogs {
		assert.Equal(t, v.BuildID, gotLogs[k].BuildID)
		assert.Equal(t, v.LogID, gotLogs[k].LogID)
		assert.Equal(t, v.Message, gotLogs[k].Message)
		assert.Equal(t, v.StepID, gotLogs[k].StepID)
	}
}

func TestServerStoppingAndRestarting(t *testing.T) {
	server := launchServer(t)
	assert.True(t, server.IsRunning())
	assert.NoError(t, server.GracefulStop())
	assert.False(t, server.IsRunning())

	go server.Serve()
	assert.True(t, server.WaitUntilRunningWithTimeout(2*time.Second))

	go server.Serve() // forceful restart
	assert.True(t, server.WaitUntilRunningWithTimeout(2*time.Second))
	assert.NoError(t, server.ForceStop())
	assert.False(t, server.IsRunning())
}

func launchServer(t *testing.T) workerserver.Server {
	server := workerserver.NewRPCServer(serverBindAddress, &mockStore{})
	go server.Serve()
	assert.True(t, server.WaitUntilRunningWithTimeout(2*time.Second))
	return server
}

func launchClient() (*workerclient.RPCClient, error) {
	client := workerclient.NewRPCClient(clientTargetAddress)
	err := client.Open()
	return client, err
}
