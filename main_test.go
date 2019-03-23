package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(t *testing.T) {
	t.Run("no args", func(t *testing.T) {
		args = []string{}
		assert.Panics(t, main)
	})
	t.Run("-list with templates", func(t *testing.T) {
		args = []string{"-debug", "-list", "python", "go"}
		assert.NotPanics(t, main)
	})
	t.Run("-list with no templates", func(t *testing.T) {
		args = []string{"-list", "-debug"}
		assert.NotPanics(t, main)
	})
	t.Run("-dump", func(t *testing.T) {
		args = []string{"-dump", "Go"}
		assert.NotPanics(t, main)
		// TODO: check stdout for expected output
	})
}
