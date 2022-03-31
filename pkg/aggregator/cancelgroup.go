package aggregator

import (
	"context"
	"sync"
)

type cancelGroup struct {
	funcs []func(ctx context.Context) error
}

func (c cancelGroup) add(f func(ctx context.Context) error) {
	c.funcs = append(c.funcs, f)
}

func (c cancelGroup) runInParallelFailFast(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	var finalErr error
	var mutex sync.Mutex
	var wg sync.WaitGroup
	wg.Add(len(c.funcs))
	for _, f := range c.funcs {
		go func(f func(ctx context.Context) error) {
			err := f(ctx)

			if err != nil {
				mutex.Lock()
				if finalErr == nil {
					finalErr = err
				}
				mutex.Unlock()
				cancel()
			}
		}(f)
	}
	wg.Wait()
	cancel()
	return finalErr
}
