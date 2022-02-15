package provisioner

import (
	"context"

	"github.com/iver-wharf/wharf-core/pkg/logger"
)

var log = logger.NewScoped("PROVISIONER")

// Provisioner is an interface declaring what methods are required
// for a provisioner.
type Provisioner interface {
	CreateWorker(ctx context.Context) (Worker, error)
	ListWorkers(ctx context.Context) (WorkerList, error)
	DeleteWorker(ctx context.Context, workerID string) error
}
