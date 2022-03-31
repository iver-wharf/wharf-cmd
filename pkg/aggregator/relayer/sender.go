package relayer

import "errors"

type sender[sent any, response any] interface {
	Send(data sent) error
	CloseAndRecv() (response, error)
}

// AsSender tries to convert the result of f to a sender with the specified
// type signature.
func AsSender[sent any, response any](f func() (any, error)) (sender[sent, response], error) {
	v, err := f()
	if err != nil {
		return nil, err
	}

	sender, ok := v.(sender[sent, response])
	if !ok {
		return nil, errors.New("can't convert to sender")
	}

	return sender, nil
}
