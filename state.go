package state

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/aphistic/gomol"

	"github.com/demosdemon/golang-app-framework/app"
)

var (
	// ErrActionRequired is returned when no command action is provided
	ErrActionRequired = errors.New("need an action")
	// ErrInvalidRepo is returned if the repo provided on the command line does not look like <owner>/<name>.
	ErrInvalidRepo = errors.New("invalid repo")
)

// The State of the application.
type State struct {
	*app.App

	// command-line flags
	debug     bool
	repo      string
	timeout   time.Duration
	action    string
	templates []string
}

func (s *State) ParseArguments() error {
	fs := flag.NewFlagSet("update-gitignore", flag.ContinueOnError)
	fs.SetOutput(s.Stderr)
	fs.Usage = usage(fs)

	debug := fs.Bool("debug", false, "print debug statements to STDERR")
	repo := fs.String("repo", "github/gitignore", "the template repository to use")
	timeout := fs.Duration("timeout", time.Second*30, "the max duration for network requests (0 for no timeout)")

	if err := fs.Parse(s.Arguments); err != nil {
		return err
	}

	s.SetDebug(*debug)
	s.SetRepo(*repo)
	s.SetTimeout(*timeout)

	args := fs.Args()
	if len(args) == 0 {
		fs.Usage()
		return ErrActionRequired
	}

	s.action = args[0]
	s.templates = args[1:]
	return nil
}

func (s *State) SetDebug(debug bool) {
	logger := s.Logger()
	s.debug = debug

	if s.debug {
		logger.SetLogLevel(gomol.LevelDebug)
	} else {
		logger.SetLogLevel(gomol.LevelInfo)
	}
}

func (s *State) Debug() bool {
	return s.debug
}

func (s *State) SetRepo(repo string) {
	s.repo = repo
}

func (s *State) Repo() string {
	return s.repo
}

func (s *State) SetTimeout(timeout time.Duration) {
	if timeout < 0 {
		timeout = 0
	}

	s.timeout = timeout
}

func (s *State) Timeout() time.Duration {
	return s.timeout
}

func (s *State) Command() (Command, error) {
	switch s.action {
	case "dump":
		return (*dumpCommand)(s), nil
	case "list":
		return (*listCommand)(s), nil
	default:
		return nil, fmt.Errorf("unrecognized action %s", s.action)
	}
}

func (s *State) Client() (*Client, error) {
	slice := strings.SplitN(s.repo, "/", 2)
	if len(slice) != 2 {
		return nil, ErrInvalidRepo
	}

	cl := &Client{
		state: s,
		owner: slice[0],
		repo:  slice[1],
	}
	cl.SetHTTPClient(nil)

	return cl, nil
}

func (s *State) deadline() (context.Context, context.CancelFunc) {
	if s.timeout > 0 {
		return context.WithTimeout(s.Context, s.timeout)
	}

	return s.Context, func() {}
}

func usage(flagset *flag.FlagSet) func() {
	return func() {
		fmt.Fprintln(flagset.Output(), `usage: update-gitignore [{flags}] {action} [{template}...]
Actions:
  dump - dumps the selected template(s) to STDOUT
  list - lists the available templates, optionally filtered by the provided arguments

{flags}    - Command line flags (see below)
{template} - The Template to dump (required for "dump") or a search string to filter (optional for "list")

Examples:
  update-gitignore list go
  update-gitignore -debug dump Go > .gitignore

Flags:`)
		flagset.PrintDefaults()
	}
}
