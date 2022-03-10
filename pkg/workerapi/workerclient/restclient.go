package workerclient

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"

	"github.com/iver-wharf/wharf-core/pkg/problem"
)

type restClient struct {
	client *http.Client
}

func newRestClient(opts Options) (*restClient, error) {
	rootCAs, err := x509.SystemCertPool()
	if err != nil && !opts.InsecureSkipVerify {
		return nil, fmt.Errorf("load system cert pool: %w", err)
	}
	if opts.InsecureSkipVerify {
		log.Warn().Message("Client is running without cert verification, this is insecure and should" +
			" not be done in production.")
		rootCAs = nil
	}
	return &restClient{
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: opts.InsecureSkipVerify,
					RootCAs:            rootCAs,
				},
			},
		},
	}, nil
}

func (c *restClient) get(url string) (*http.Response, error) {
	return c.client.Get(url)
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