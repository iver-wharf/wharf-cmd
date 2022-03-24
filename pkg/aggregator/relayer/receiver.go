package relayer

import (
	"errors"
)

type receiver[received any] interface {
	Recv() (received, error)
	CloseSend() error
}

// AsReceiver tries to convert the result of f to a receiver with the specified
// type signature.
func AsReceiver[received any](f func() (any, error)) (receiver[received], error) {
	v, err := f()
	if err != nil {
		return nil, err
	}

	receiver, ok := v.(receiver[received])
	if !ok {
		return nil, errors.New("can't convert to receiver")
	}

	return receiver, nil
}
