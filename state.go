package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/google/go-github/v24/github"
	"golang.org/x/oauth2"
)

// The State of the application.
type State struct {
	// command line arguments
	Debug     bool
	Dump      bool
	List      bool
	Templates []string

	// calculated from `repo`
	Owner string
	Repo  string

	logger *log.Logger
	client *github.Client
	ctx    context.Context
}

// NewState build a new application state object
func NewState() *State {
	prog := path.Base(os.Args[0])
	state := &State{
		logger: log.New(os.Stderr, prog+" ", log.LstdFlags|log.Llongfile),
		ctx:    context.Background(),
	}

	flag.Usage = state.Usage
	flag.BoolVar(&state.Debug, "debug", false, "Print debug statements to STDERR.")
	flag.BoolVar(&state.Dump, "dump", false, "Dump the specified templates to STDOUT.")
	flag.BoolVar(&state.List, "list", false, "List the available templates. If any, arguments are used to filter the results.")
	repo := flag.String("repo", "github/gitignore", "The template repo to use.")

	flag.Parse()
	state.Templates = flag.Args()

	slice := strings.SplitN(*repo, "/", 2)
	if len(slice) != 2 {
		state.Fatal("Invalid repo: %v.", slice)
	}
	state.Owner = slice[0]
	state.Repo = slice[1]

	state.Log("Debug = %t", state.Debug)
	state.Log("Dump  = %t", state.Dump)
	state.Log("List  = %t", state.List)
	state.Log("Owner = %s", state.Owner)
	state.Log("Repo  = %s", state.Repo)
	state.Log("Templates = %v", state.Templates)

	if state.Dump && state.List {
		state.Fatal("-dump and -list are mutually exclusive.")
	}

	if !state.Dump && !state.List {
		state.Fatal("one of -dump or -list is required.")
	}

	if state.Dump && len(state.Templates) == 0 {
		state.Fatal("Must provide at least one template with -dump.")
	}

	return state
}

// Log prints the formatted message if `s.Debug` is true.
func (s *State) Log(format string, v ...interface{}) {
	if s.Debug {
		s.logger.Output(2, fmt.Sprintf(format, v...))
	}
}

// Panic prints the formatted message and calls `panic` with the `err`.
func (s *State) Panic(err error, format string, v ...interface{}) {
	s.logger.Output(2, fmt.Sprintf(format, v...))
	panic(err)
}

// Fatal prints the formatted message and calls `os.Exit`
func (s *State) Fatal(format string, v ...interface{}) {
	s.logger.Output(2, fmt.Sprintf(format, v...))
	os.Exit(1)
}

// Client fetches the cached GitHub client
//
// If the environment variable GITHUB_TOKEN is found, uses it to authenticate
// GitHub API requests.
func (s *State) Client() *github.Client {
	if s.client == nil {
		token, found := os.LookupEnv("GITHUB_TOKEN")
		if found {
			ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
			tc := oauth2.NewClient(s.ctx, ts)
			s.client = github.NewClient(tc)
		} else {
			s.client = github.NewClient(nil)
		}
	}
	return s.client
}

// Usage prints command line usage information.
func (State) Usage() {
	fmt.Printf("usage: %s [flags] [template ...]\n", os.Args[0])
	flag.PrintDefaults()
}

// GetRepository fetches the GitHub repository.
func (s *State) GetRepository() (*github.Repository, error) {
	rv, _, err := s.Client().Repositories.Get(s.ctx, s.Owner, s.Repo)
	return rv, err
}

// GetBranch fetches the specified branch from the GitHub repository.
func (s *State) GetBranch(branch string) (*github.Branch, error) {
	rv, _, err := s.Client().Repositories.GetBranch(s.ctx, s.Owner, s.Repo, branch)
	return rv, err
}

// GetTree fetches the specified tree from the GitHub repository.
func (s *State) GetTree(sha string) (*github.Tree, error) {
	rv, _, err := s.Client().Git.GetTree(s.ctx, s.Owner, s.Repo, sha, false)
	return rv, err
}
