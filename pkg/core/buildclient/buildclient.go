package buildclient

import (
	"os"
	"time"

	"github.com/iver-wharf/wharf-api-client-go/pkg/wharfapi"
)

type Client interface {
	PostLog(buildID uint, message string) error
	PutStatus(buildID uint, status wharfapi.BuildStatus) error
	PostLogWithStatus(buildID uint, message string, status wharfapi.BuildStatus) error
}

type client struct {
	wharfClient wharfapi.Client
}

func New(authHeader string) Client {
	return client{wharfClient: newWharfClient(authHeader)}
}

func newWharfClient(authHeader string) wharfapi.Client {
	return wharfapi.Client{
		ApiUrl:     os.Getenv("WHARF_API_URL"),
		AuthHeader: authHeader,
	}
}

func (p client) PostLog(buildID uint, message string) error {
	return p.wharfClient.PostLog(buildID, wharfapi.BuildLog{
		Message:   message,
		Timestamp: time.Now().UTC(),
	})
}

func (p client) PutStatus(buildID uint, status wharfapi.BuildStatus) error {
	_, err := p.wharfClient.PutStatus(buildID, status)
	return err
}

func (p client) PostLogWithStatus(buildID uint, message string, status wharfapi.BuildStatus) error {
	err := p.PutStatus(buildID, status)
	if err != nil {
		return err
	}

	return p.PostLog(buildID, message)
}
