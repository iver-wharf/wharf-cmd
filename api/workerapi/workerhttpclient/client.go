package workerhttpclient

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/iver-wharf/wharf-cmd/api/workerapi/workerhttpserver/model/response"
)

type workerClient struct {
	targetAddress string
	targetPort    string

	client http.Client
}

func NewClient(targetAddress, targetPort string) Client {
	rootCAs, err := x509.SystemCertPool()
	if err != nil {
		panic(err)
	}

	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: rootCAs,
			},
		},
	}

	return &workerClient{
		targetAddress: targetAddress,
		targetPort:    targetPort,
		client:        client,
	}
}

func (c *workerClient) GetBuildSteps() ([]response.Step, error) {
	res, err := c.client.Get(fmt.Sprintf("https://%s:%s/api/build/step", c.targetAddress, c.targetPort))
	if err != nil {
		panic(err)
	}

	bytes, err := ioutil.ReadAll(res.Body)

	var steps []response.Step
	if err := json.Unmarshal(bytes, &steps); err != nil {
		panic(err)
	}

	return steps, nil
}
