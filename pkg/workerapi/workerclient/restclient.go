package workerclient

import (
	"context"
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

func (c *restClient) do(ctx context.Context, method, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, err
	}
	return c.client.Do(req)
}

func (c *restClient) delete(ctx context.Context, url string) (*http.Response, error) {
	return c.do(ctx, http.MethodDelete, url)
}

func (c *restClient) get(ctx context.Context, url string) (*http.Response, error) {
	return c.do(ctx, http.MethodGet, url)
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
