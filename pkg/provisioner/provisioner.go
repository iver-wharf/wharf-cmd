package provisioner

import (
	"context"
)

// Provisioner is an interface declaring what methods are required
// for a provisioner.
type Provisioner interface {
	CreateWorker(ctx context.Context) (Worker, error)
	ListWorkers(ctx context.Context) ([]Worker, error)
	DeleteWorker(ctx context.Context, workerID string) error
}
