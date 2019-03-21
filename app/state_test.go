package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewState(t *testing.T) {
	defer PanicOnError(InitLogging())

	assert.Panics(t, func() {
		NewState([]string{"-repo", "invalid"})
	})
	assert.Panics(t, func() {
		NewState([]string{"-dump", "-list", "-repo=github/gitignore"})
	})
	assert.Panics(t, func() {
		NewState([]string{"-debug", "-dump", "-repo=github/notreal"})
	})
	assert.Panics(t, func() {
		NewState([]string{"-repo", "also/invalid"})
	})
}
