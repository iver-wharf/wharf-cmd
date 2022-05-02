package provisioner

import (
	"context"
)

// Provisioner is an interface declaring what methods are required
// for a provisioner.
type Provisioner interface {
	CreateWorker(ctx context.Context, args WorkerArgs) (Worker, error)
	ListWorkers(ctx context.Context) ([]Worker, error)
	DeleteWorker(ctx context.Context, workerID string) error
}

// WorkerArgs are arguments fed to the wharf-cmd-worker.
type WorkerArgs struct {
	GitCloneURL    string
	GitCloneBranch string
	SubDir         string

	Environment string
	Stage       string

	Inputs         map[string]any
	AdditionalVars map[string]any

	WharfInstanceID string
	ProjectID       uint
	BuildID         uint
}
