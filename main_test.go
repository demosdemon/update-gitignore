package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(t *testing.T) {
	t.Run("-list", func(t *testing.T) {
		args = []string{"-list", "python"}
		assert.NotPanics(t, main)
	})
}
