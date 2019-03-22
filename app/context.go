package app

import (
	"context"
	"sync"

	"github.com/aphistic/gomol"
)

// PanicUnlessCancelled should be deferred immediately and silently captures
// panics if the context has been canceled.
func PanicUnlessCanceled(ctx context.Context) {
	if p := recover(); p != nil {
		if err, ok := p.(error); ok {
			select {
			case <-ctx.Done():
				// only swallow the panic if it was sent by the context
				if err == ctx.Err() {
					gomol.Debugf("swallowing error: %v", err)
					return
				}
			default:
				// context isn't marked done
			}
		}
		// repanic
		panic(p)
	}
}

type Task = func(ctx context.Context) (err error)

type ContextPool struct {
	sync.WaitGroup

	ctx    context.Context
	cancel context.CancelFunc
	result chan error
}

func NewContextPool(ctx context.Context) *ContextPool {
	ctx, cancel := context.WithCancel(ctx)
	return &ContextPool{
		ctx:    ctx,
		cancel: cancel,
		result: make(chan error, 1),
	}
}

func (p *ContextPool) Cancel() error {
	p.cancel()
	for err := range p.result {
		if err == nil {
			continue
		}
		return err
	}
	return nil
}

func (p *ContextPool) AddTask(task Task) {
	p.Add(1)
	go p.executeTask(task)
}

func (p *ContextPool) executeTask(task Task) {
	defer p.Done()
	defer func() {
		if v := recover(); v != nil {
			p.result <- NewWrappedPanic(v)
			close(p.result)
		}
	}()

	if err := task(p.ctx); err != nil {
		select {
		case p.result <- err:
		case <-p.ctx.Done():
			p.result <- p.ctx.Err()
			close(p.result)
		}
	}
}
