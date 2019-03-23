package app

import (
	"context"
	"errors"
	"flag"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/go-github/v24/github"
	"github.com/gosuri/uitable"
	"golang.org/x/oauth2"
)

var (
	// ErrInvalidRepo is returned when the command line argument for `-repo` does not match the pattern
	// "<owner>/<repository>"
	ErrInvalidRepo = errors.New("invalid repo")
	// ErrInvalidTimeout is returned when the command line argument for `-timeout` is less than zero. Zero is valid and
	// indicates no timeout
	ErrInvalidTimeout = errors.New("invalid timeout")
	// ErrMutuallyExclusiveOption is returned when both `-list` and `-dump` are provided in the command line arguments
	ErrMutuallyExclusiveOption = errors.New("-list and -dump are mutually exclusive")
	// ErrActionRequired is returned when neither `-list` or `-dump` are provided in the command line arguments
	ErrActionRequired = errors.New("one of -list or -dump is required")
	// ErrActionArguments is returned when `-dump` is specified with no additional arguments
	ErrActionArguments = errors.New("-dump requires at least one argument")
)

// The State of the application.
type State struct {
	// Debug holds a boolean value depicting whether or not `-debug` was provided as a command line argument
	Debug bool
	// Dump holds a boolean value depicting whether or not `-dump` was provided as a command line argument
	Dump bool
	// List holds a boolean value depicting whether or not `-list` was provided as a command line argument
	List bool
	// Templates is a string slice containing any additional command line arguments
	Templates []string

	// Owner holds the parsed owner value from the `-repo` command line flag
	Owner string
	// Repo holds the parsed repository value from the `-repo` command line flag
	Repo string

	// Context holds the executing context for the application. Context may have a deadline associated if `-timeout` was
	// provided on the command line
	Context context.Context
	// Cancel holds the function necessary to cancel the context early, if so desired
	Cancel context.CancelFunc

	mu     sync.Mutex
	client *github.Client
	token  *oauth2.Token
}

// NewState builds a new application state object, attaches the supplied context, and parses the supplied command line
// arguments
func NewState(ctx context.Context, arguments []string, output io.Writer) (*State, error) {
	fs := flag.NewFlagSet("update-gitignore", flag.ContinueOnError)
	fs.SetOutput(output)

	debug := fs.Bool("debug", false, "print debug statements to STDERR")
	dump := fs.Bool("dump", false, "dump the specified templates to STDOUT")
	list := fs.Bool(
		"list",
		false,
		"list the available templates; if any, provided arguments are used to filter the results",
	)
	repo := fs.String("repo", "github/gitignore", "the template repository to use")
	timeout := fs.Duration("timeout", time.Second*30, "the max runtime duration, set to 0 for no timeout")

	if err := fs.Parse(arguments); err != nil {
		return nil, err
	}

	slice := strings.SplitN(*repo, "/", 2)
	if len(slice) < 2 {
		return nil, ErrInvalidRepo
	}

	if *timeout < 0 {
		return nil, ErrInvalidTimeout
	}

	var cancel context.CancelFunc

	if *timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, *timeout)
	} else {
		ctx, cancel = context.WithCancel(ctx)
	}

	var token *oauth2.Token

	if tok, ok := os.LookupEnv("GITHUB_TOKEN"); ok {
		token = &oauth2.Token{AccessToken: tok}
	}

	state := &State{
		Debug:     *debug,
		Dump:      *dump,
		List:      *list,
		Templates: fs.Args(),
		Owner:     slice[0],
		Repo:      slice[1],
		Context:   ctx,
		Cancel:    cancel,
		token:     token,
	}

	return state, state.Validate(*repo, *timeout)
}

func (s *State) Validate(repo string, timeout time.Duration) error {
	if s.Dump && s.List {
		return ErrMutuallyExclusiveOption
	}

	if !s.Dump && !s.List {
		return ErrActionRequired
	}

	if s.Dump && len(s.Templates) < 1 {
		return ErrActionArguments
	}

	return nil
}

// PrintDebug uses the supplied `print` function to write tabelized debugging information
func (s *State) PrintDebug(print func(...interface{}) error) error {
	if err := print("===> State"); err != nil {
		return err
	}

	table := uitable.New()
	table.AddRow("Debug", s.Debug)
	table.AddRow("Dump", s.Dump)
	table.AddRow("List", s.List)
	table.AddRow("Templates", strings.Join(s.Templates, ", "))
	table.AddRow("Owner", s.Owner)
	table.AddRow("Repo", s.Repo)
	table.AddRow("Context", s.Context)

	if err := print(table); err != nil {
		return err
	}

	return nil
}

func (s *State) SetToken(token *oauth2.Token) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client != nil {
		panic("a client has already been created with the previous key")
	}

	s.token = token
}

func (s *State) Token() (*oauth2.Token, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.token == nil {
		return nil, errors.New("no token")
	}

	return s.token, nil
}

func (s *State) Client() *github.Client {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client == nil {
		var httpClient *http.Client
		if s.token != nil {
			httpClient = oauth2.NewClient(s.Context, s)
		}
		s.client = github.NewClient(httpClient)
	}
	return s.client
}
