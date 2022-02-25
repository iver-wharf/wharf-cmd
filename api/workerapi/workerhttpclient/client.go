package workerhttpclient

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/iver-wharf/wharf-cmd/api/workerapi/workerhttpserver/model/response"
	"github.com/iver-wharf/wharf-core/pkg/cacertutil"
)

type workerClient struct {
	address string
	port    string
	client  *http.Client
}

func NewClient(address, port string) (Client, error) {
	client, err := cacertutil.NewHTTPClientWithCerts("/etc/iver-wharf/wharf-cmd/localhost.crt")
	if err != nil {
		return nil, err
	}
	return &workerClient{
		address: address,
		port:    port,
		client:  client,
	}, nil
}

func (c *workerClient) GetBuildSteps() ([]response.Step, error) {
	res, err := c.client.Get(fmt.Sprintf("http://%s:%s/api/build/step", c.address, c.port))
	if err != nil {
		return nil, err
	}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var steps []response.Step
	if err := json.Unmarshal(bytes, &steps); err != nil {
		return nil, err
	}
	return steps, nil
}
