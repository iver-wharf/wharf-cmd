package aggregator

// Aggregator aggregates.
type Aggregator interface {
	Serve() error
}
