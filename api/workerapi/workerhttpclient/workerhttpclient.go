package workerhttpclient

import "github.com/iver-wharf/wharf-cmd/api/workerapi/workerhttpserver/model/response"

type Client interface {
	GetBuildSteps() ([]response.Step, error)
}
