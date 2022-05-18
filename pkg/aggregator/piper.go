package aggregator

import (
	"github.com/iver-wharf/wharf-api-client-go/v2/pkg/wharfapi"
	v1 "github.com/iver-wharf/wharf-cmd/api/workerapi/v1"
	"github.com/iver-wharf/wharf-cmd/pkg/workerapi/workerclient"
)

// Piper pipes.
type Piper interface {
	PipeMessage() error // Returns io.EOF when done.
	Close() error
}

type artifactEventsPiper struct {
	wharfapi wharfapi.Client
	worker   workerclient.Client
	in       v1.Worker_StreamArtifactEventsClient
}
