package main

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/demosdemon/update-gitignore/app"
)

func TestMain(t *testing.T) {
	cases := []struct {
		name        string
		arguments   []string
		environment []string
		stdin       string
		stdout      string
		stderr      string
		exitcode    int
	}{
		{
			"plain",
			[]string{"list"},
			[]string{},
			"",
			"",
			"",
			0,
		},
		{
			"with error",
			[]string{},
			[]string{},
			"",
			"",
			"",
			2,
		},
	}

	// cannot run these in parallel
	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			instance = &app.App{
				Arguments:   tt.arguments,
				Environment: tt.environment,
				Context:     context.Background(),
				Stdin:       strings.NewReader(tt.stdin),
				Stdout:      new(bytes.Buffer),
				Stderr:      new(bytes.Buffer),
				Exit: func(code int) {
					panic(code)
				},
			}

			assert.PanicsWithValue(t, tt.exitcode, main)
		})
	}
}
