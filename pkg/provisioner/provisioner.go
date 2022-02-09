package provisioner

import (
	"context"

	"github.com/iver-wharf/wharf-core/pkg/logger"
	v1 "k8s.io/api/core/v1"
)

var log = logger.NewScoped("PROVISIONER")

// Provisioner is an interface declaring what methods are required
// for a provisioner.
type Provisioner interface {
	CreateWorker(ctx context.Context) (*v1.Pod, error)
	ListWorkers(ctx context.Context) ([]v1.Pod, error)
	DeleteWorker(ctx context.Context, workerID string) error
}
