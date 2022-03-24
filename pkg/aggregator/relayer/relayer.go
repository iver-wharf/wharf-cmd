package relayer

// Relayer is an interface used for sending received data to another location,
// converting it in the process.
type Relayer[received any, sent any, response any] interface {
	// Relay receives data from a source, and sends it to a destination.
	//
	// Returns multiple error messages as a string array, which is nil if no
	// errors occurred.
	Relay() []string
}

type convert[received any, sent any] func(v received) sent

// New creates a new relayer that relays data from the receiver to the sender.
func New[received any, sent any, response any](r receiver[received], s sender[sent, response], c convert[received, sent]) Relayer[received, sent, response] {
	return &grpcRelayer[received, sent, response]{
		receiver: r,
		sender:   s,
		convert:  c,
	}
}