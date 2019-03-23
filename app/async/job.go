package async

import (
	"context"
	"errors"
	"sync"
)

var (
	ErrNilContext  = errors.New("nil context")
	ErrInvalidTask = errors.New("invalid task")
)

type Job struct {
	cancel context.CancelFunc
	ch     chan error

	mu     sync.Mutex
	result error
}

func New(ctx context.Context, task Task) (*Job, error) {
	if ctx == nil {
		return nil, ErrNilContext
	}

	if task == nil {
		return nil, ErrInvalidTask
	}

	ctx, cancel := context.WithCancel(ctx)
	job := &Job{cancel: cancel, ch: make(chan error)}

	go func() {
		defer RecoverFromPanic(job.ch)
		job.ch <- task(ctx)
	}()

	return job, nil
}

func (j *Job) Result() error {
	j.mu.Lock()
	defer j.mu.Unlock()

	result, ok := <-j.ch
	if ok {
		j.result = result
	} else {
		result = j.result
	}

	return result
}

func (j *Job) Cancel() error {
	ch := make(chan error)
	go func() {
		defer RecoverFromPanic(ch)
		j.cancel()
		ch <- j.Result()
	}()

	return <-ch
}
