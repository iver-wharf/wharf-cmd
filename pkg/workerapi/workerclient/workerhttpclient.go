package workerclient

import (
	"crypto/tls"
	"crypto/x509"
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

// NewClient creates a client that can communicate with a worker HTTP server.
//
// Uses the system cert pool, and optionally allows skipping cert verification,
//
// Note that skipping verification is insecure and should not be done in a
// production environment!
func NewClient(address string, insecureSkipVerify bool) (HTTPClient, error) {
	rootCAs, err := x509.SystemCertPool()
	if err != nil && !insecureSkipVerify {
		log.Debug().Message("Getting system cert pool failed, and insecure skip verify is false.")
		return nil, err
	}
	if insecureSkipVerify {
		log.Warn().Message("Client is running without cert verification, this is insecure and should" +
			" not be done in production.")
		rootCAs = nil
	}
	return &workerHTTPClient{
		address: address,
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: insecureSkipVerify,
					RootCAs:            rootCAs,
				},
			},
		},
	}, nil
}

// NewClientWithCerts creates a client that can communicate with a Worker HTTP server.
//
// Appends certs at the provided path to the system cert pool and uses that.
func NewClientWithCerts(address, certFilePath string) (HTTPClient, error) {
	client, err := cacertutil.NewHTTPClientWithCerts(certFilePath)
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
	if res == nil {
		return err
	}
	if problem.IsHTTPResponse(res) {
		prob, parseErr := problem.ParseHTTPResponse(res)
		if parseErr == nil {
			return prob
		}
		return parseErr
	}
	return err
}
