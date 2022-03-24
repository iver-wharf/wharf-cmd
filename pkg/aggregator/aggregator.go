package aggregator

// Aggregator pulls data from workers and sends them to Wharf API.
type Aggregator interface {
	Serve() error
}
