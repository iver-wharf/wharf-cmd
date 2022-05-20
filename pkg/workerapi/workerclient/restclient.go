package workerclient

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
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

func (c *restClient) get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return c.client.Do(req)
}

func assertResponseOK(res *http.Response, err error) error {
	if err != nil {
		return err
	}
	if res == nil {
		return errors.New("response is nil")
	}
	if problem.IsHTTPResponse(res) {
		prob, parseErr := problem.ParseHTTPResponse(res)
		if parseErr != nil {
			return parseErr
		}
		return prob
	}
	return nil
}
