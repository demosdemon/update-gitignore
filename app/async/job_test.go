package async

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func theAnswer(out chan<- int) func(context.Context) error {
	return func(ctx context.Context) error {
		select {
		case out <- 42:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func TestJob(t *testing.T) {
	out := make(chan int, 1)
	job, err := New(context.Background(), theAnswer(out))
	if assert.NoError(t, err) && assert.NotNil(t, job) {
		err := job.Result()
		if assert.NoError(t, err) {
			x := <-out
			assert.EqualValues(t, 42, x)
		}
	}
}

func TestJobNilContext(t *testing.T) {
	out := make(chan int)
	job, err := New(nil, theAnswer(out)) // nolint
	if assert.Nil(t, job) && assert.EqualError(t, err, "nil context") {
		select {
		case <-out:
			assert.Fail(t, "out should block")
		default:
		}
	}
}

func TestJobNilTask(t *testing.T) {
	job, err := New(context.Background(), nil)
	assert.EqualError(t, err, "invalid task")
	assert.Nil(t, job)
}

func TestJobWithError(t *testing.T) {
	task := func(ctx context.Context) error {
		return assert.AnError
	}

	job, err := New(context.Background(), task)
	if assert.NotNil(t, job) && assert.NoError(t, err) {
		err = job.Result()
		assert.Equal(t, assert.AnError, err)
	}
}

func TestJobThatPanics(t *testing.T) {
	out := make(chan int)
	close(out)

	// send on a closed channel causes a run-time panic
	job, err := New(context.Background(), theAnswer(out))
	if assert.NotNil(t, job) && assert.NoError(t, err) {
		err := job.Result()
		assert.Error(t, err)

		_, ok := err.(*WrappedPanic)
		assert.True(t, ok)
	}
}

func TestJobWithDeadline(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// send on a nil channel blocks forever
	job, err := New(ctx, theAnswer(nil))
	if assert.NotNil(t, job) && assert.NoError(t, err) {
		err := job.Result()
		assert.EqualError(t, err, "context deadline exceeded")
	}
}

func TestJobRunsOnlyOnce(t *testing.T) {
	var count int32
	task := func(ctx context.Context) error {
		return fmt.Errorf("count %d", atomic.AddInt32(&count, 1))
	}

	job, err := New(context.Background(), task)
	if assert.NotNil(t, job) && assert.NoError(t, err) {
		e1 := job.Result()
		assert.EqualError(t, e1, "count 1")
		e2 := job.Result()
		assert.EqualError(t, e2, "count 1")
		assert.Equal(t, e1, e2)
	}
}

func TestJobIsCancelable(t *testing.T) {
	out := make(chan int)

	job, err := New(context.Background(), theAnswer(out))
	if assert.NotNil(t, job) && assert.NoError(t, err) {
		err := job.Cancel()
		assert.EqualError(t, err, "context canceled")
		assert.Equal(t, err, job.Result())
		assert.NotPanics(t, func() {
			err2 := job.Cancel()
			assert.Equal(t, err, err2)
		})
	}
}
