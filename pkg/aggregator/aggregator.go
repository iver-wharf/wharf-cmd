package aggregator

import (
	"context"
)

// Aggregator pulls data from workers and sends them to Wharf API.
type Aggregator interface {
	Serve(ctx context.Context) error
}

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
