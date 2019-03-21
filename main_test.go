package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(t *testing.T) {
	t.Run("-list", func(t *testing.T) {
		args = []string{"-debug", "-list"}
		assert.NotPanics(t, main)
	})
}
