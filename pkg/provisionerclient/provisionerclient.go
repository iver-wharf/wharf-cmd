package provisionerclient

import (
	"errors"

	"github.com/iver-wharf/wharf-cmd/pkg/provisioner"
)

type Client struct{}

func (c Client) ListWorkers() ([]provisioner.Worker, error) {
	return nil, errors.New("not implemented")
}
