package app

import (
	"context"

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
