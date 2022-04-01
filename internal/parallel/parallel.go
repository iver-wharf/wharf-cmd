package parallel

import (
	"context"
	"sync"
)

// Func is a function declaration used in parallel runs.
type Func func(ctx context.Context) error

// Group is a list of functions to run in parallel.
type Group []Func

// AddFunc adds a function to the group to later be used in the parallel call.
func (g *Group) AddFunc(f Func) {
	*g = append(*g, f)
}

// RunFailFast will run all functions in parallel in separate goroutines, and
// will cancel them all as soon as one of them returns an error. The resulting
// error is the error from the first function that errors.
func (g *Group) RunFailFast(ctx context.Context) error {
	var rg groupRun
	return rg.runAll(ctx, *g)
}

type groupRun struct {
	mutex    sync.Mutex
	finalErr error
	wg       sync.WaitGroup
}

func (rg *groupRun) runAll(ctx context.Context, group Group) error {
	ctx, cancel := context.WithCancel(ctx)
	rg.wg.Add(len(group))
	for _, f := range group {
		go rg.runFunc(ctx, cancel, f)
	}
	rg.wg.Wait()

	// It's still guaranteed that cancel() to be called in runOnce,
	// but this shuts up the compiler warnings.
	// It's safe to call cancel() multiple times.
	cancel()
	return rg.finalErr
}

func (rg *groupRun) runFunc(ctx context.Context, cancel context.CancelFunc, f Func) {
	defer rg.wg.Done()
	err := f(ctx)
	if err != nil {
		rg.mutex.Lock()
		if rg.finalErr == nil {
			rg.finalErr = err
		}
		rg.mutex.Unlock()
		cancel()
	}
}
