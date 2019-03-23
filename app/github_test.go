package app_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/demosdemon/update-gitignore/app"
)

func TestTree(t *testing.T) {
	defer app.PanicOnError(app.InitLogging())

	t.Run("Basics", func(t *testing.T) {
		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, time.Minute*5)
		defer cancel()

		state, err := app.NewState(ctx, []string{"-list"}, os.Stdout)
		assert.NoError(t, err)
		ch := state.Tree(ctx)
		for x := range ch {
			assert.Contains(t, x.Path, ".gitignore")
			assert.NotEqual(t, "", x.Name)
			if len(x.Tags) > 0 {
				assert.NotEqual(t, "", x.Tags[0])
			}
		}
	})

	t.Run("Cancelable", func(t *testing.T) {
		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)

		state, err := app.NewState(ctx, []string{"-list"}, os.Stdout)
		assert.NoError(t, err)
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
	})
}

func TestGetDefaultBranch(t *testing.T) {
	defer app.PanicOnError(app.InitLogging())

	ctx := context.Background()
	state, err := app.NewState(ctx, []string{"-dump", "Python"}, os.Stdout)
	assert.NoError(t, err)

	t.Run("cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(ctx)
		cancel()

		branch := state.GetDefaultBranch(ctx)
		assert.EqualValues(t, "", branch)
	})
	t.Run("not cancelled", func(t *testing.T) {
		branch := state.GetDefaultBranch(ctx)
		assert.EqualValues(t, "master", branch)
	})
	t.Run("invalid repo", func(t *testing.T) {
		state, err := app.NewState(ctx, []string{"-list", "-repo=demosdemon/thisrepodoesnotexist"}, os.Stdout)
		assert.NoError(t, err)
		assert.Panics(t, func() {
			_ = state.GetDefaultBranch(ctx)
			panic(assert.AnError) // TODO: fix test
		})
	})
}

func TestGetBranchHead(t *testing.T) {
	defer app.PanicOnError(app.InitLogging())

	ctx := context.Background()

	t.Run("cancelled", func(t *testing.T) {
		state, err := app.NewState(ctx, []string{"-debug", "-list"}, os.Stdout)
		assert.NoError(t, err)
		ctx, cancel := context.WithCancel(ctx)
		cancel()

		commit := state.GetBranchHead(ctx, "master")
		assert.EqualValues(t, "", commit)
	})
	t.Run("not cancelled", func(t *testing.T) {
		// intentionally defunct repo in attempt to make the sha constant
		state, err := app.NewState(ctx, []string{"-list", "-repo=demosdemon/CheckBuyvm"}, os.Stdout)
		assert.NoError(t, err)
		commit := state.GetBranchHead(ctx, "master")
		assert.EqualValues(t, "251502fe2ce94571548baf1710cde2beca037d57", commit)
	})
	t.Run("invalid repo", func(t *testing.T) {
		state, err := app.NewState(ctx, []string{"-list", "-repo=demosdemon/thisrepodoesnotexist"}, os.Stdout)
		assert.NoError(t, err)
		assert.Panics(t, func() {
			_ = state.GetBranchHead(ctx, "master")
			panic(assert.AnError) // TODO: fix test
		})
	})
}
