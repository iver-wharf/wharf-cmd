package workerclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/iver-wharf/wharf-cmd/pkg/workerapi/workerserver/model/response"
	"github.com/iver-wharf/wharf-core/pkg/cacertutil"
	"github.com/iver-wharf/wharf-core/pkg/problem"
)

type workerHTTPClient struct {
	address string
	client  *http.Client
}

// NewClient creates a client that can communicate with a Worker HTTP server.
func NewClient(address string) (HTTPClient, error) {
	client, err := cacertutil.NewHTTPClientWithCerts("/etc/iver-wharf/wharf-cmd/localhost.crt")
	if err != nil {
		return nil, err
	}
	return &workerHTTPClient{
		address: address,
		client:  client,
	}, nil
}

func (c *workerHTTPClient) ListBuildSteps() ([]response.Step, error) {
	res, err := c.client.Get(fmt.Sprintf("http://%s/api/build/step", c.address))
	if err := errorIfBad(res, err); err != nil {
		return nil, err
	}

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Error().WithError(err).Message("Failed closing response body reader.")
		}
	}()

	var steps []response.Step
	if err := json.Unmarshal(bytes, &steps); err != nil {
		return nil, err
	}

	return steps, nil
}

func (c *workerHTTPClient) ListArtifacts() ([]response.Artifact, error) {
	res, err := c.client.Get(fmt.Sprintf("http://%s/api/artifact", c.address))
	if err := errorIfBad(res, err); err != nil {
		return nil, err
	}

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Error().WithError(err).Message("Failed closing response body reader.")
		}
	}()

	var artifacts []response.Artifact
	if err := json.Unmarshal(bytes, &artifacts); err != nil {
		return nil, err
	}

	return artifacts, nil
}

func (c *workerHTTPClient) DownloadArtifact(artifactID uint) (io.ReadCloser, error) {
	res, err := c.client.Get(fmt.Sprintf("http://%s/api/artifact/%d/download", c.address, artifactID))
	if err := errorIfBad(res, err); err != nil {
		return nil, err
	}
	return res.Body, nil
}

func errorIfBad(res *http.Response, err error) error {
	if problem.IsHTTPResponse(res) {
		prob, parseErr := problem.ParseHTTPResponse(res)
		if parseErr == nil {
			return prob
		}
		return parseErr
	}
	return err
}
