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
			"invalid",
			[]string{"invalid"},
			[]string{},
			"",
			"",
			"[\x1b[31mERROR\x1b[0m] unrecognized action invalid {\"filename\":\"base.go\",\"lineno\":488,\"seq\":1}\n",
			2,
		},
		{
			"with error",
			[]string{},
			[]string{},
			"",
			"",
			"usage: update-gitignore [{flags}] {action} [{template}...]\nActions:\n  dump - dumps the selected template(s) to STDOUT\n  list - lists the available templates, optionally filtered by the provided arguments\n\n{flags}    - Command line flags (see below)\n{template} - The Template to dump (required for \"dump\") or a search string to filter (optional for \"list\")\n\nExamples:\n  update-gitignore list go\n  update-gitignore -debug dump Go > .gitignore\n\nFlags:\n  -debug\n    \tprint debug statements to STDERR\n  -repo string\n    \tthe template repository to use (default \"github/gitignore\")\n  -timeout duration\n    \tthe max duration for network requests (0 for no timeout) (default 30s)\n[\x1b[31mERROR\x1b[0m] need an action {\"filename\":\"base.go\",\"lineno\":488,\"seq\":1}\n",
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
			stdout := instance.Stdout.(*bytes.Buffer)
			assert.Equal(t, tt.stdout, stdout.String())
			stderr := instance.Stderr.(*bytes.Buffer)
			assert.Equal(t, tt.stderr, stderr.String())
		})
	}
}
