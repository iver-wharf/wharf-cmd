package aggregator

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/iver-wharf/wharf-cmd/pkg/workerapi/workerclient"
)

type artifactSource struct {
	ctx    context.Context
	worker workerclient.Client
}

var _ Source[*workerclient.ArtifactEvent] = artifactSource{}

func (s artifactSource) PushInto(dst chan<- *workerclient.ArtifactEvent) error {
	reader, err := s.worker.StreamArtifactEvents(s.ctx, &workerclient.ArtifactEventsRequest{})
	if err != nil {
		return fmt.Errorf("open artifact events stream from wharf-cmd-worker: %w", err)
	}
	defer reader.CloseSend()
	for {
		artifactEvent, err := reader.Recv()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return err
			}
			break
		}
		dst <- artifactEvent
	}
	return nil
}
