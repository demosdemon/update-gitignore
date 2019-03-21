package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewState(t *testing.T) {
	t.Run("Invalid Repo", func(t *testing.T) {
		assert.Panics(t, func() {
			NewState([]string{"-repo", "invalid"})
		})
	})
	t.Run("Mutual Exclusion", func(t *testing.T) {
		assert.Panics(t, func() {
			NewState([]string{"-dump", "list"})
		})
	})
	t.Run("Required Argument", func(t *testing.T) {
		assert.Panics(t, func() {
			NewState([]string{"-debug", "-dump"})
		})
	})
}
