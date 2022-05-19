package aggregator

import (
	"context"
	"errors"
	"fmt"

	"github.com/iver-wharf/wharf-api-client-go/v2/pkg/wharfapi"
	v1 "github.com/iver-wharf/wharf-cmd/api/workerapi/v1"
	"github.com/iver-wharf/wharf-cmd/pkg/workerapi/workerclient"
)

type artifactEventsPiper struct {
	wharfapi wharfapi.Client
	worker   workerclient.Client
	in       v1.Worker_StreamArtifactEventsClient
}

func newArtifactEventsPiper(ctx context.Context, wharfapi wharfapi.Client, worker workerclient.Client) (artifactEventsPiper, error) {
	in, err := worker.StreamArtifactEvents(ctx, &workerclient.ArtifactEventsRequest{})
	if err != nil {
		return artifactEventsPiper{}, fmt.Errorf("open status events stream from wharf-cmd-worker: %w", err)
	}
	return artifactEventsPiper{
		wharfapi: wharfapi,
		worker:   worker,
		in:       in,
	}, nil
}

func (p artifactEventsPiper) PipeMessage() error {
	msg, err := p.readArtifactEvent()
	if err != nil {
		return fmt.Errorf("read artifact event: %w", err)
	}
	if err := p.writeArtifactEvent(msg); err != nil {
		return fmt.Errorf("write artifact event: %w", err)
	}
	return nil
}

// Close closes all active streams.
func (p artifactEventsPiper) Close() error {
	if err := p.in.CloseSend(); err != nil {
		log.Error().
			WithError(err).
			Message("Failed closing artifact events stream from worker.")
	}
	return nil
}

func (p artifactEventsPiper) readArtifactEvent() (*workerclient.ArtifactEvent, error) {
	if p.in == nil {
		return nil, errors.New("input stream is nil")
	}
	artifactEvent, err := p.in.Recv()
	if err != nil {
		return nil, err
	}
	return artifactEvent, nil
}

func (p artifactEventsPiper) writeArtifactEvent(artifactEvent *workerclient.ArtifactEvent) error {
	// No way to send to wharf DB through stream currently
	// so we're just logging it here.
	log.Debug().
		WithUint64("step", artifactEvent.StepID).
		WithString("name", artifactEvent.Name).
		WithUint64("id", artifactEvent.ArtifactID).
		Message("Received artifact event.")
	return nil
}
