package aggregator

import "io"

// PipeCloser pipes data, and should always be closed when done with it.
type PipeCloser interface {
	// PipeMessage should pipe a single message, and when done should return
	// io.EOF to indicate success.
	PipeMessage() error

	io.Closer
}
