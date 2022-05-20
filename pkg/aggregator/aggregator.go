package aggregator

import (
	"context"
)

// Aggregator pulls data from workers and sends them to Wharf API.
type Aggregator interface {
	Serve(ctx context.Context) error
}
