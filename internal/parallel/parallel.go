package parallel

import (
	"context"
	"fmt"
	"sync"
)

// Func is a function declaration used in parallel runs.
type Func func(ctx context.Context) error

type task struct {
	name string
	f    Func
}

// Group is a list of functions to run in parallel.
type Group []task

// AddFunc adds a function to the group to later be used in the parallel call.
// The name is prepended to the error message, if any.
func (g *Group) AddFunc(name string, f Func) {
	*g = append(*g, task{name, f})
}

// RunCancelEarly will run all functions in parallel in separate goroutines, and
// will cancel them all as soon as one of them returns an error. The resulting
// error is the error from the first function that errors.
func (g *Group) RunCancelEarly(ctx context.Context) error {
	return runAll(ctx, *g)
}

type groupRun struct {
	mutex    sync.Mutex
	finalErr error
	wg       sync.WaitGroup
}

func runAll(ctx context.Context, group Group) error {
	var gr groupRun
	ctx, cancel := context.WithCancel(ctx)
	gr.wg.Add(len(group))
	for _, t := range group {
		go gr.runFunc(ctx, cancel, t)
	}
	gr.wg.Wait()

	// It's still guaranteed that cancel() to be called in runOnce,
	// but this shuts up the compiler warnings.
	// It's safe to call cancel() multiple times.
	cancel()
	return gr.finalErr
}

func (gr *groupRun) runFunc(ctx context.Context, cancel context.CancelFunc, t task) {
	defer gr.wg.Done()
	err := t.f(ctx)
	if err != nil {
		gr.mutex.Lock()
		if gr.finalErr == nil {
			gr.finalErr = fmt.Errorf("%s: %w", t.name, err)
		}
		gr.mutex.Unlock()
		cancel()
	}
}
