package aggregator

import (
	"context"
)

// Aggregator pulls data from workers and sends them to Wharf API.
type Aggregator interface {
	Serve(ctx context.Context) error
}

// relay takes a source and relays values from it by passing them to the
// provided function.
//
// If the provided function returns an error, relaying will stop and the error
// will be returned as-is.
func relay[T any](src Source[T], relayFunc func(v T) error) error {
	dst := make(chan T)
	go func() {
		src.PushInto(dst)
		close(dst)
	}()
	for v, ok := <-dst; ok; v, ok = <-dst {
		if err := relayFunc(v); err != nil {
			return err
		}
	}
	return nil
}
