package state

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	usageValue = chain(
		"usage: update-gitignore [{flags}] {action} [{template}...]\n",
		"Actions:\n",
		"  dump - dumps the selected template(s) to STDOUT\n",
		"  list - lists the available templates, optionally filtered by the provided arguments\n",
		"\n",
		"{flags}    - Command line flags (see below)\n",
		"{template} - The Template to dump (required for \"dump\") ",
		"or a search string to filter (optional for \"list\")\n",
		"\n",
		"Examples:\n",
		"  update-gitignore list go\n",
		"  update-gitignore -debug dump Go > .gitignore\n",
		"\n",
		"Flags:\n",
		usageLine("-debug", "print debug statements to STDERR"),
		usageLine("-repo string", "the template repository to use (default \"github/gitignore\")"),
		usageLine("-timeout duration", "the max duration for network requests (0 for no timeout) (default 30s)"),
	)
)

func chain(v ...string) string {
	w := strings.Builder{}
	for _, s := range v {
		w.WriteString(s)
	}
	return w.String()
}

func usageLine(flag, help string) string {
	return fmt.Sprintf("  %s\n    \t%s\n", flag, help)
}

func TestState_ParseArguments(t *testing.T) {
	type expected struct {
		err     *string
		stdout  string
		stderr  string
		debug   bool
		repo    string
		timeout time.Duration
	}

	type testcase struct {
		name     string
		state    *State
		expected expected
	}

	readBuffer := func(w io.Writer) string {
		b, _ := w.(*bytes.Buffer)
		buf, err := ioutil.ReadAll(b)
		if err != nil {
			panic(err)
		}

		return string(buf)
	}

	cases := []testcase{
		{
			"no arguments",
			&State{App: newApp(nil)},
			expected{
				strptr("need an action"),
				"",
				usageValue,
				false,
				"github/gitignore",
				time.Second * 30,
			},
		},
		{
			"invalid flag",
			&State{App: newApp(nil, "--not-for-you")},
			expected{
				strptr("flag provided but not defined: -not-for-you"),
				"",
				chain(
					"flag provided but not defined: -not-for-you\n",
					usageValue,
				),
				false,
				"",
				0,
			},
		},
		{
			"debug flag",
			&State{App: newApp(nil, "-debug", "list")},
			expected{
				nil,
				"",
				"",
				true,
				"github/gitignore",
				time.Second * 30,
			},
		},
		{
			"extreme timeout",
			&State{App: newApp(nil, "-timeout", "30m", "list")},
			expected{
				nil,
				"",
				"",
				false,
				"github/gitignore",
				time.Minute * 30,
			},
		},
		{
			"negative timeout",
			&State{App: newApp(nil, "-timeout", "-30s", "list")},
			expected{
				nil,
				"",
				"",
				false,
				"github/gitignore",
				0,
			},
		},
		{
			"custom repo",
			&State{App: newApp(nil, "-repo", "demosdemon/gitignore", "list")},
			expected{
				nil,
				"",
				"",
				false,
				"demosdemon/gitignore",
				time.Second * 30,
			},
		},
		{
			"invalid repo",
			&State{App: newApp(nil, "-repo=invalid", "list")},
			expected{
				nil,
				"",
				"",
				false,
				"invalid",
				time.Second * 30,
			},
		},
	}

	t.Parallel()
	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			defer tt.state.Logger().ShutdownLoggers()
			err := tt.state.ParseArguments()
			errEquals(t, tt.expected.err, err)
			assert.Equal(t, tt.expected.stdout, readBuffer(tt.state.Stdout))
			assert.Equal(t, tt.expected.stderr, readBuffer(tt.state.Stderr))
			assert.Equal(t, tt.expected.debug, tt.state.Debug())
			assert.Equal(t, tt.expected.repo, tt.state.Repo())
			assert.Equal(t, tt.expected.timeout, tt.state.Timeout())
		})
	}
}

func TestState_Command(t *testing.T) {
	cases := []struct {
		name  string
		state *State
		err   *string
	}{
		{
			"dump",
			&State{App: newApp(nil, "dump")},
			nil,
		},
		{
			"list",
			&State{App: newApp(nil, "list")},
			nil,
		},
		{
			"invalid",
			&State{App: newApp(nil, "invalid")},
			strptr("unrecognized action invalid"),
		},
	}

	t.Parallel()
	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			defer tt.state.Logger().ShutdownLoggers()
			err := tt.state.ParseArguments()
			assert.NoError(t, err)

			cmd, err := tt.state.Command()
			errEquals(t, tt.err, err)

			if cmd != nil {
				name := cmd.GetName()
				assert.Equal(t, tt.name, name)

				rv := cmd.Run()
				assert.Equal(t, ExitStatus(0), rv)
			}
		})
	}
}

func TestState_Client(t *testing.T) {
	cases := []struct {
		name  string
		state *State
		err   *string
	}{
		{
			"valid",
			&State{App: newApp(nil, "test")},
			nil,
		},
		{
			"invalid",
			&State{App: newApp(nil, "-repo", "invalid", "test")},
			strptr("invalid repo"),
		},
	}

	t.Parallel()
	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			defer tt.state.Logger().ShutdownLoggers()
			err := tt.state.ParseArguments()
			assert.NoError(t, err)
			_, err = tt.state.Client()
			errEquals(t, tt.err, err)
		})
	}
}

func TestState_deadline(t *testing.T) {
	a := newApp(nil, "-timeout", "1ns", "test")
	defer a.Logger().ShutdownLoggers()
	s := State{App: a}
	_ = s.ParseArguments()
	c, _ := s.Client()
	c.SetHTTPClient(&http.Client{Transport: newReplay("anonymous")})
	_, err := c.GetBlob("0000000000000000000000000000000000000000")
	assert.EqualError(t, err, "context deadline exceeded")
}
