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

			state := State{}
			client := state.Client(ctx)

			rl, _, err := client.RateLimits(ctx)
			assert.NoError(t, err)
			assert.Truef(t, rl.Core.Limit >= 5000, "rl.Core.Limit < 5000: %d", rl.Core.Limit)
		})
	})
}

func TestTree(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Minute*5)
	defer cancel()

	state := NewState([]string{"-list"})
	ch := state.Tree(ctx)
	for x := range ch {
		assert.Contains(t, x.Path, ".gitignore")
		assert.NotEqual(t, "", x.Name)
		if len(x.Tags) > 0 {
			assert.NotEqual(t, "", x.Tags[0])
		}
	}
}

func TestTreeCancelable(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	state := NewState([]string{"-list"})
	ch := state.Tree(ctx)

	x, ok := <-ch
	assert.True(t, ok)
	assert.NotNil(t, x)
	cancel()

	i := 0
	for range ch {
		i++
	}

	assert.True(t, 5 <= i || i <= 6)
}

func TestTreeUnrealisticTimeout(t *testing.T) {
	ctx := context.Background()
	state := NewState([]string{"-debug", "-list"})
	branch := state.GetDefaultBranch(ctx)
	commit := state.GetBranchHead(ctx, branch)

	ctx, cancel := context.WithTimeout(ctx, time.Microsecond)
	defer cancel()
	ch := state.getTree(ctx, commit)
	_, ok := <-ch
	assert.False(t, ok)
}

func TestGetDefaultBranch(t *testing.T) {
	state := NewState([]string{"-dump", "Python"})

	t.Run("cancelled", func(t *testing.T) {
		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)
		cancel()

		branch := state.GetDefaultBranch(ctx)
		assert.EqualValues(t, "", branch)
	})
	t.Run("not cancelled", func(t *testing.T) {
		ctx := context.Background()
		branch := state.GetDefaultBranch(ctx)
		assert.EqualValues(t, "master", branch)
	})
	t.Run("invalid repo", func(t *testing.T) {
		ctx := context.Background()
		state := State{Owner: "demosdemon", Repo: "thisrepodoesnotexist"}

		assert.Panics(t, func() {
			_ = state.GetDefaultBranch(ctx)
		})
	})
}

func TestGetBranchHead(t *testing.T) {
	t.Run("cancelled", func(t *testing.T) {
		state := NewState([]string{"-debug", "-list"})
		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)
		cancel()

		commit := state.GetBranchHead(ctx, "master")
		assert.EqualValues(t, "", commit)
	})
	t.Run("not cancelled", func(t *testing.T) {
		ctx := context.Background()
		// intentionally defunct repo in attempt to make the sha constant
		state := State{Owner: "demosdemon", Repo: "CheckBuyvm"}

		commit := state.GetBranchHead(ctx, "master")
		assert.EqualValues(t, "251502fe2ce94571548baf1710cde2beca037d57", commit)
	})
	t.Run("invalid repo", func(t *testing.T) {
		ctx := context.Background()
		state := State{Owner: "demosdemon", Repo: "thisrepodoesnotexist"}

		assert.Panics(t, func() {
			_ = state.GetBranchHead(ctx, "master")
		})
	})
}
