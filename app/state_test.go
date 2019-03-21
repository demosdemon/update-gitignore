package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewState(t *testing.T) {
	assert.Panics(t, func() {
		NewState([]string{"-repo", "invalid"})
	})
	assert.Panics(t, func() {
		NewState([]string{"-dump", "list", "-repo=github/gitignore"})
	})
	assert.Panics(t, func() {
		NewState([]string{"-debug", "-dump", "-repo=github/notreal"})
	})
}
