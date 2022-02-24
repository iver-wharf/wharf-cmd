package workerhttpclient

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/iver-wharf/wharf-cmd/api/workerapi/workerhttpserver/model/response"
	"github.com/iver-wharf/wharf-cmd/pkg/config"
	"github.com/iver-wharf/wharf-core/pkg/cacertutil"
)

var cfg config.WorkerClientConfig

func init() {
	c, err := config.LoadConfig()
	if err != nil {
		fmt.Println("Failed to read config:", err)
		os.Exit(1)
	}
	cfg = c.Worker.Client
}

type workerClient struct {
	client http.Client
}

func NewClient() (Client, error) {
	client, err := cacertutil.NewHTTPClientWithCerts(cfg.CA.CertsFile)
	if err != nil {
		return nil, err
	}
	return &workerClient{
		client: *client,
	}, nil
}

func (c *workerClient) GetBuildSteps() ([]response.Step, error) {
	res, err := c.client.Get(fmt.Sprintf("https://%s/api/build/step", cfg.HTTP.BindAddress))
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
