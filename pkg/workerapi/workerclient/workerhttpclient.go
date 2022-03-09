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
	opts    ClientOptions
}

// ClientOptions contains options that can be used in the creation
// of a new client.
type ClientOptions struct {
	// InsecureSkipVerify disables cert verification if set to true.
	//
	// Should NOT be true in a production environment.
	InsecureSkipVerify bool
}

// NewHTTPClient creates a client that can communicate with a worker HTTP server.
//
// The address should include the host scheme, e.g.:
//   http://
//   https://
//
// Uses the system cert pool, and optionally allows skipping cert verification,
//
// Note that skipping verification is insecure and should not be done in a
// production environment!
func NewHTTPClient(address string, opts ClientOptions) (HTTPClient, error) {
	rootCAs, err := x509.SystemCertPool()
	if err != nil && !opts.InsecureSkipVerify {
		return nil, fmt.Errorf("load system cert pool: %w", err)
	}
	if opts.InsecureSkipVerify {
		log.Warn().Message("Client is running without cert verification, this is insecure and should" +
			" not be done in production.")
		rootCAs = nil
	}
	return &workerHTTPClient{
		address: address,
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: opts.InsecureSkipVerify,
					RootCAs:            rootCAs,
				},
			},
		},
		opts: opts,
	}, nil
}

// NewClientWithCerts creates a client that can communicate with a Worker HTTP server.
//
// The address should include the host scheme, e.g.:
//   http://
//   https://
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

func (c *workerHTTPClient) ListBuildSteps() (steps []response.Step, finalError error) {
	res, err := c.client.Get(fmt.Sprintf("%s/api/build/step", c.address))
	if err := assertResponseOK(res, err); err != nil {
		finalError = err
		return
	}

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		finalError = err
		return
	}
	defer closeAndSetError(res.Body, &finalError)
	if err := json.Unmarshal(bytes, &steps); err != nil {
		finalError = err
		return
	}
	finalError = err
	return
}

func (c *workerHTTPClient) ListArtifacts() (artifacts []response.Artifact, finalError error) {
	res, err := c.client.Get(fmt.Sprintf("%s/api/artifact", c.address))
	if err := assertResponseOK(res, err); err != nil {
		finalError = err
		return
	}

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		finalError = err
		return
	}
	defer closeAndSetError(res.Body, &finalError)
	if err := json.Unmarshal(bytes, &artifacts); err != nil {
		finalError = err
		return
	}

	return
}

func (c *workerHTTPClient) DownloadArtifact(artifactID uint) (io.ReadCloser, error) {
	res, err := c.client.Get(fmt.Sprintf("%s/api/artifact/%d/download", c.address, artifactID))
	if err := assertResponseOK(res, err); err != nil {
		return nil, err
	}
	return res.Body, nil
}

func assertResponseOK(res *http.Response, err error) error {
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

// closeAndSetError may be used to set the named return variable inside a
// deferred call if that deferred call failed.
//
// NOTE: The errPtr argument must be a pointer to a named result parameters,
// otherwise it will not affect the calling function's returned value.
func closeAndSetError(closer io.Closer, errPtr *error) {
	closeErr := closer.Close()
	if errPtr != nil && *errPtr == nil {
		*errPtr = closeErr
	}
}
