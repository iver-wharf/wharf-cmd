package provisionerclient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/iver-wharf/wharf-cmd/pkg/provisioner"
	"github.com/iver-wharf/wharf-core/pkg/logger"
	"github.com/iver-wharf/wharf-core/pkg/problem"
)

// Client is a HTTP client that talks to wharf-cmd-provisioner.
type Client struct {
	// APIURL is the base API URL used. Example value:
	// 	http://wharf-cmd-provisioner.default.svc.cluster.local
	APIURL string
}

var log = logger.NewScoped("PROVISIONER-CLIENT")

// ListWorkers returns a slice of all workers.
func (c Client) ListWorkers() ([]provisioner.Worker, error) {
	u, err := buildURL(c.APIURL, "api", "worker")
	if err != nil {
		return nil, err
	}
	resp, err := doRequest(http.MethodGet, u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	var workers []provisioner.Worker
	if err := dec.Decode(&workers); err != nil {
		return nil, err
	}
	return workers, nil
}

// DeleteWorker will terminate a worker based on its ID.
func (c Client) DeleteWorker(workerID string) error {
	u, err := buildURL(c.APIURL, "api", "worker", workerID)
	if err != nil {
		return err
	}
	resp, err := doRequest(http.MethodDelete, u)
	if err != nil {
		return err
	}
	return resp.Body.Close()
}

// Ping pongs.
func (c Client) Ping() error {
	u, err := buildURL(c.APIURL)
	if err != nil {
		return err
	}
	_, err = doRequest(http.MethodGet, u)
	return err
}

func doRequest(method string, u *url.URL) (*http.Response, error) {
	urlStr := u.String()
	req, err := http.NewRequest(method, urlStr, nil)
	if err != nil {
		return nil, err
	}
	log.Debug().
		WithString("method", method).
		WithString("url", urlStr).
		Message("")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if err := parseErrorResponse(resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func parseErrorResponse(resp *http.Response) error {
	if problem.IsHTTPResponse(resp) {
		prob, err := problem.ParseHTTPResponse(resp)
		if err != nil {
			return err
		}
		return prob
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("non-2xx status code: %s", resp.Status)
	}
	return nil
}

func buildURL(base string, paths ...string) (*url.URL, error) {
	u, err := url.Parse(base)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "" {
		return nil, fmt.Errorf("URL is missing scheme: %q", base)
	}
	if u.Host == "" {
		return nil, fmt.Errorf("URL is missing host: %q", base)
	}
	var pathBuilder strings.Builder
	var rawPathBuilder strings.Builder
	if u.Path != "" {
		pathBuilder.WriteString(strings.TrimSuffix(u.Path, "/"))
		rawPathBuilder.WriteString(strings.TrimSuffix(u.EscapedPath(), "/"))
	}
	for _, segment := range paths {
		pathBuilder.WriteByte('/')
		rawPathBuilder.WriteByte('/')
		pathBuilder.WriteString(segment)
		rawPathBuilder.WriteString(url.PathEscape(segment))
	}
	u.Path = pathBuilder.String()
	u.RawPath = rawPathBuilder.String()
	return u, nil
}
