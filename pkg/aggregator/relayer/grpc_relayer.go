package relayer

import (
	"errors"
	"io"

	"github.com/iver-wharf/wharf-core/pkg/logger"
)

var log = logger.NewScoped("RELAY")

type grpcRelayer[received any, sent any, response any] struct {
	receiver[received]
	sender[sent, response]
	convert[received, sent]

	errs []string
	used bool
}

func (r *grpcRelayer[fromWorker, fromWharf, toWharf]) Relay() []string {
	if r.used {
		return []string{"can only relay once"}
	}

	for {
		received, ok := r.recv()
		if !ok {
			break
		}
		if ok := r.send(received); !ok {
			break
		}
	}
	r.close()
	r.used = true
	return r.errs
}

func (r *grpcRelayer[received, sent, response]) close() {
	if err := r.CloseSend(); err != nil {
		r.errs = append(r.errs, err.Error())
	}

	_, err := r.CloseAndRecv()
	if err != nil {
		r.errs = append(r.errs, err.Error())
	}
}

func (r *grpcRelayer[received, sent, response]) recv() (received, bool) {
	v, err := r.Recv()
	if err != nil {
		if !errors.Is(err, io.EOF) {
			r.errs = append(r.errs, err.Error())
		}
		return *new(received), false
	}
	return v, true
}

func (r *grpcRelayer[received, sent, response]) send(v received) bool {
	if err := r.Send(r.convert(v)); err != nil {
		r.errs = append(r.errs, err.Error())
		return false
	}
	return true
}
