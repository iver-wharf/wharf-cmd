package closer

import (
	"io"

	"gopkg.in/typ.v3/pkg/chans"
)

type chanCloser[C chans.Sender[E], E any] struct {
	ch C
}

// Close implements io.Closer.
func (c chanCloser[C, E]) Close() error {
	close(c.ch)
	return nil
}

// NewChanCloser returns an io.Closer that closes
func NewChanCloser[C chans.Sender[E], E any](ch C) io.Closer {
	return chanCloser[C, E]{ch}
}
