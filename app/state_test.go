package app_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"

	"github.com/demosdemon/update-gitignore/app"
)

var (
	ctx = context.Background()
)

type newStateTestCase struct {
	args []string

	valid     bool
	timeout   time.Duration
	errString string
	stdout    string
	stderr    string
}

var tests = []newStateTestCase{
	{
		[]string{"--not-a-valid-flag", "list"},
		false,
		0,
		"flag provided but not defined: -not-a-valid-flag",
		"",
		"",
	},
	{
		[]string{"-repo=invalid", "list"},
		false,
		0,
		"invalid repo",
		"",
		"",
	},
	{
		[]string{"-timeout", "-30s", "list"},
		false,
		0,
		"invalid timeout",
		"",
		"",
	},
	{
		[]string{"list"},
		true,
		30 * time.Second,
		"",
		"",
		"",
	},
	{
		[]string{"-timeout", "0", "list"},
		true,
		0,
		"",
		"",
		"",
	},
	{
		[]string{"-timeout", "60m", "list"},
		true,
		time.Hour,
		"",
		"",
		"",
	},
	{
		[]string{},
		false,
		0,
		"an action, one of dump or list, is required",
		"",
		"",
	},
	{
		[]string{"dump"},
		true,
		30 * time.Second,
		"",
		"",
		"",
	},
}

func newState(args ...string) (state *app.State, stdout string, stderr string, err error) {
	fd0 := strings.NewReader("")
	fd1 := new(strings.Builder)
	fd2 := new(strings.Builder)

	state, err = app.New(ctx, args, fd0, fd1, fd2)
	if err != nil {
		return state, stdout, stderr, err
	}

	stdout = fd1.String()
	stderr = fd2.String()
	return state, stdout, stderr, err
}

func TestNewState(t *testing.T) {
	for _, c := range tests {
		state, stdout, stderr, err := newState(c.args...)
		if c.valid {
			assert.NotNilf(t, state, "%#v", c)
			assert.NoErrorf(t, err, "%#v", c)
			assert.Equalf(t, c.timeout, state.Timeout(), "%#v", c)
		} else {
			assert.Nilf(t, state, "%#v", c)
			assert.EqualErrorf(t, err, c.errString, "%#v", c)
		}

		assert.Equalf(t, c.stdout, stdout, "%#v", c)
		assert.Equalf(t, c.stderr, stderr, "%#v", c)
	}
}

func TestPrintDebug(t *testing.T) {
	state, _, _, _ := newState("-repo", "demosdemon/thisdoesnotexist", "list", "go", "python")

	var w strings.Builder
	print := func(v ...interface{}) error {
		_, err := fmt.Fprintln(&w, v...)
		return err
	}

	err := state.PrintDebug(print)
	assert.NoError(t, err)
	assert.Equal(
		t,
		chain(
			"===> State\n",
			"Owner    \tdemosdemon        \n",
			"Repo     \tthisdoesnotexist  \n",
			"Context  \tcontext.Background\n",
			"Timeout  \t30s               \n",
			"Action   \tlist              \n",
			"Arguments\tgo python         \n",
			"Client   \t<nil>             \n",
		),
		w.String(),
	)
}

func TestPrintDebugFaultyPrinter(t *testing.T) {
	state, _, _, _ := newState("list")

	var w strings.Builder
	print := func(...interface{}) error {
		return errors.New("always an error")
	}

	err := state.PrintDebug(print)
	assert.EqualError(t, err, "always an error")
	assert.Equal(t, "", w.String())

	count := 0
	print = func(v ...interface{}) error {
		count++
		if count == 1 {
			_, err := fmt.Fprintln(&w, v...)
			return err
		}
		return fmt.Errorf("error %d", count)
	}

	err = state.PrintDebug(print)
	assert.EqualError(t, err, "error 2")
	assert.Equal(t, "===> State\n", w.String())
}

func TestSetTokenPanicsAfterClientCreation(t *testing.T) {
	state, _, _, _ := newState("list")
	client := state.Client()
	assert.NotNil(t, client)

	assert.PanicsWithValue(
		t,
		"a client has already been created with the previous token",
		func() {
			state.SetToken(new(oauth2.Token))
		},
	)
}

func TestSetToken(t *testing.T) {
	state, _, _, _ := newState("list")
	assert.NotPanics(t, func() { state.SetToken(nil) })
	assert.NotPanics(t, func() { state.SetToken(&oauth2.Token{}) })
	assert.NotPanics(t, func() { state.SetToken(nil) })
}

func TestClientWithNoEnvironment(t *testing.T) {
	state, _, _, _ := newState("list")
	state.SetToken(nil)

	client := state.Client()
	assert.NotNil(t, client)

	rl, _, err := client.RateLimits(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 60, rl.Core.Limit)
}

func TestClientWithEnvironmentToken(t *testing.T) {
	token, ok := os.LookupEnv("GITHUB_TOKEN")
	assert.True(t, ok, "Missing environment variable GITHUB_TOKEN")

	state, _, _, _ := newState("list")
	state.SetToken(&oauth2.Token{AccessToken: token})

	tok, err := state.Token()
	assert.NotNil(t, tok)
	assert.NoError(t, err)
	assert.Equal(t, token, tok.AccessToken)

	client := state.Client()
	assert.NotNil(t, client)

	rl, _, err := client.RateLimits(ctx)
	assert.NoError(t, err)
	assert.Truef(t, rl.Core.Limit >= 5000, "Rate Limit < 5000: %d", rl.Core.Limit)
}

func chain(v ...string) string {
	w := strings.Builder{}
	for _, s := range v {
		w.WriteString(s)
	}
	return w.String()
}
