package app_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/demosdemon/update-gitignore/app"
)

var ctx = context.Background()
var usage = chain(
	"Usage of update-gitignore:\n",
	usageLine(
		"-debug",
		"print debug statements to STDERR",
	),
	usageLine(
		"-dump",
		"dump the specified templates to STDOUT",
	),
	usageLine(
		"-list",
		chain(
			"list the available templates; if any, provided arguments are used to filter ",
			"the results",
		),
	),
	usageLine(
		"-repo string",
		"the template repository to use (default \"github/gitignore\")",
	),
	usageLine(
		"-timeout duration",
		"the max runtime duration, set to 0 for no timeout (default 30s)",
	),
)

func TestNewStateInvalidArguments(t *testing.T) {
	w := strings.Builder{}
	state, err := app.NewState(context.Background(), []string{"--not-a-valid-flag"}, &w)
	assert.Nil(t, state)
	msg := "flag provided but not defined: -not-a-valid-flag"
	assert.EqualError(t, err, msg)
	assert.EqualValues(t, msg+"\n"+usage, w.String())
}

func TestNewStateInvalidRepo(t *testing.T) {
	w := strings.Builder{}
	state, err := app.NewState(ctx, []string{"-list", "-repo=invalid"}, &w)
	assert.Nil(t, state)
	assert.EqualError(t, err, "invalid repo")
	assert.Equal(t, "", w.String())
}

func TestNewStateInvalidTimeout(t *testing.T) {
	w := strings.Builder{}
	state, err := app.NewState(ctx, []string{"-list", "-timeout", "-30s"}, &w)
	assert.Nil(t, state)
	assert.EqualError(t, err, "invalid timeout")
	assert.Equal(t, "", w.String())

}

func TestNewStateNoTimeoutArg(t *testing.T) {
	expected := time.Now().Add(30 * time.Second)
	w := strings.Builder{}
	state, err := app.NewState(ctx, []string{"-list"}, &w)
	assert.NotNil(t, state)
	assert.NoError(t, err)
	assert.Equal(t, "", w.String())
	deadline, ok := state.Context.Deadline()
	assert.True(t, ok)
	assert.WithinDuration(t, expected, deadline, time.Millisecond)
}

func TestNewStateNoTimeout(t *testing.T) {
	w := strings.Builder{}
	state, err := app.NewState(ctx, []string{"-list", "-timeout=0"}, &w)
	assert.NotNil(t, state)
	assert.NoError(t, err)
	assert.Equal(t, "", w.String())
	_, ok := state.Context.Deadline()
	assert.False(t, ok)
	assert.Contains(t, fmt.Sprintf("%v", state.Context), "context.Background.WithCancel")
}

func TestNewStateHighTimeout(t *testing.T) {
	expected := time.Now().Add(time.Hour)
	w := strings.Builder{}
	state, err := app.NewState(ctx, []string{"-list", "-timeout", "60m"}, &w)
	assert.NotNil(t, state)
	assert.NoError(t, err)
	assert.Equal(t, "", w.String())
	deadline, ok := state.Context.Deadline()
	assert.True(t, ok)
	assert.WithinDuration(t, expected, deadline, time.Millisecond)

}

func TestNewStateNoArguments(t *testing.T) {
	w := strings.Builder{}
	state, err := app.NewState(ctx, []string{}, &w)
	assert.Nil(t, state)
	assert.EqualError(t, err, "one of -list or -dump is required")
	assert.Equal(t, "", w.String())

}

func TestNewStateWithBothListAndDump(t *testing.T) {
	w := strings.Builder{}
	state, err := app.NewState(ctx, []string{"-list", "-dump", "go"}, &w)
	assert.Nil(t, state)
	assert.EqualError(t, err, "-list and -dump are mutually exclusive")
	assert.Equal(t, "", w.String())

}

func TestNewStateWithDumpAndNoArguments(t *testing.T) {
	w := strings.Builder{}
	state, err := app.NewState(ctx, []string{"-dump"}, &w)
	assert.Nil(t, state)
	assert.EqualError(t, err, "-dump requires at least one argument")
	assert.Equal(t, "", w.String())

}

func TestPrintDebug(t *testing.T) {
	state := app.State{
		Debug:     false,
		Dump:      false,
		List:      false,
		Templates: []string{},
		Owner:     "github",
		Repo:      "gitignore",
		Context:   ctx,
		Cancel:    nil,
	}

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
			"Debug    \tfalse             \n",
			"Dump     \tfalse             \n",
			"List     \tfalse             \n",
			"Templates\t                  \n",
			"Owner    \tgithub            \n",
			"Repo     \tgitignore         \n",
			"Context  \tcontext.Background\n",
		),
		w.String(),
	)
}

func TestPrintDebugFaultyPrinter(t *testing.T) {
	state := app.State{
		Debug:     true,
		Dump:      true,
		List:      true,
		Templates: []string{"go"},
		Owner:     "github",
		Repo:      "gitignore",
		Context:   ctx,
		Cancel:    nil,
	}

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

func usageLine(flag, help string) string {
	return fmt.Sprintf("  %s\n    \t%s\n", flag, help)
}

func chain(v ...string) string {
	w := strings.Builder{}
	for _, s := range v {
		w.WriteString(s)
	}
	return w.String()
}
