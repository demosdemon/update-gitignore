package app

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func clearAndRestoreEnviron(f func()) {
	environ := os.Environ()
	os.Clearenv()

	f()
	// allow `f` to muck about with the environment
	os.Clearenv()

	for _, line := range environ {
		if line == "" {
			continue
		}

		split := strings.SplitN(line, "=", 2)
		key := split[0]
		if len(split) == 2 {
			os.Setenv(key, split[1])
		} else {
			os.Setenv(key, "")
		}
	}
}

func TestClient(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Minute*30)
	defer cancel()

	token, ok := os.LookupEnv("GITHUB_TOKEN")
	assert.True(t, ok, "Missing environment variable GITHUB_TOKEN")

	t.Run("Test Client with nil state", func(t *testing.T) {
		var state *State
		client := state.Client(ctx)
		assert.Nil(t, client)
	})

	t.Run("Test Client with no environment", func(t *testing.T) {
		clearAndRestoreEnviron(func() {
			var state = State{}
			client := state.Client(ctx)

			rl, _, err := client.RateLimits(ctx)
			assert.NoError(t, err)
			assert.Equal(t, 60, rl.Core.Limit)
		})
	})

	t.Run("Test Client with environment token", func(t *testing.T) {
		clearAndRestoreEnviron(func() {
			err := os.Setenv("GITHUB_TOKEN", token)
			assert.NoError(t, err)

			var state = State{}
			client := state.Client(ctx)

			rl, _, err := client.RateLimits(ctx)
			assert.NoError(t, err)
			assert.Truef(t, rl.Core.Limit >= 5000, "rl.Core.Limit < 5000: %d", rl.Core.Limit)
		})
	})
}
