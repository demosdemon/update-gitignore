package app

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrappedPanicIsAnError(t *testing.T) {
	var p interface{} = &WrappedPanic{}
	err, ok := p.(error)
	assert.True(t, ok)
	assert.EqualError(t, err, "panic: <nil>\n")
}

func TestRecoverFromPanic(t *testing.T) {
	ch := make(chan error, 1)
	var obj struct{}

	assert.NotPanics(t, func() {
		defer close(ch)
		defer RecoverFromPanic(ch)
		panic(obj)
	})

	err, ok := <-ch
	assert.True(t, ok)
	assert.Error(t, err)

	wp, ok := err.(*WrappedPanic)
	assert.True(t, ok)
	assert.Error(t, wp)

	assert.Equal(t, obj, wp.Value)

	lines := strings.Split(wp.Stack, "\n")
	assert.Contains(t, lines[0], "goroutine")
	assert.Contains(t, lines[0], "[running]")
	assert.Equal(t, "github.com/demosdemon/update-gitignore/app.TestRecoverFromPanic.func1()", lines[1])

	select {
	case err, ok := <-ch:
		assert.False(t, ok)
		assert.NoError(t, err)
	default:
		assert.Fail(t, "error channel was not closed")
	}
}