package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/aphistic/gomol"
	gomolconsole "github.com/aphistic/gomol-console"
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
	// ErrActionRequired is required when no command action is provided
	ErrActionRequired = errors.New("an action, one of dump or list, is required")
	// ErrActionArguments is returned when `dump` is specified with no additional arguments
	ErrActionArguments = errors.New("`dump` requires at least one argument")
)

// The State of the application.
type State struct {
	// Owner holds the parsed owner value from the `-repo` command line flag
	Owner string
	// Repo holds the parsed repository value from the `-repo` command line flag
	Repo string

	ctx       context.Context
	timeout   time.Duration
	logger    *gomol.Base
	action    string
	arguments []string
	stdin     io.Reader
	stdout    io.Writer
	stderr    io.Writer

	tokenMu sync.Mutex
	client  *github.Client
	token   *oauth2.Token
}

const (
	logTemplate = `{{.Template.Format "2006-01-02 15:04:05.000"}} [{{color}}{{ucase .LevelName}}{{reset}}] {{.Message}}`

	fullLogTemplate = logTemplate + `{{if .Attrs}} {{json .Attrs}}{{end}}`
)

// NewState builds a new application state object, attaches the supplied context, and parses the supplied command line
// arguments
func New(
	ctx context.Context,
	arguments []string,
	stdin io.Reader,
	stdout, stderr io.Writer,
) (*State, error) {
	fs := flag.NewFlagSet(path.Base(os.Args[0]), flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = usage(fs)

	debug := fs.Bool("debug", false, "print debug statements to STDERR")
	repo := fs.String("repo", "github/gitignore", "the template repository to use")
	timeout := fs.Duration("timeout", time.Second*30, "the max duration for network requests, set to 0 for no timeout")

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

	var token *oauth2.Token
	if tok, ok := os.LookupEnv("GITHUB_TOKEN"); ok {
		token = &oauth2.Token{AccessToken: tok}
	}

	args := fs.Args()
	if len(args) == 0 {
		return nil, ErrActionRequired
	}

	action, args := args[0], args[1:]

	logger, _ := newLogger(stderr)
	if *debug {
		logger.SetLogLevel(gomol.LevelDebug)
	} else {
		logger.SetLogLevel(gomol.LevelInfo)
	}

	state := &State{
		Owner: slice[0],
		Repo:  slice[1],

		ctx:       ctx,
		timeout:   *timeout,
		logger:    logger,
		action:    action,
		arguments: args,
		stdin:     stdin,
		stdout:    stdout,
		stderr:    stderr,

		token: token,
	}

	return state, nil
}

// PrintDebug uses the supplied `print` function to write tabelized debugging information
func (s *State) PrintDebug(print func(...interface{}) error) error {
	if err := print("===> State"); err != nil {
		return err
	}

	table := uitable.New()
	table.AddRow("Owner", s.Owner)
	table.AddRow("Repo", s.Repo)
	table.AddRow("Context", s.ctx)
	table.AddRow("Timeout", s.timeout)
	table.AddRow("Action", s.action)
	table.AddRow("Arguments", strings.Join(s.arguments, " "))
	table.AddRow("Client", s.client)

	if err := print(table); err != nil {
		return err
	}

	return nil
}

func (s *State) Timeout() time.Duration {
	return s.timeout
}

func (s *State) SetToken(token *oauth2.Token) {
	s.tokenMu.Lock()
	defer s.tokenMu.Unlock()

	if s.client != nil {
		panic("a client has already been created with the previous token")
	}

	s.token = token
}

func (s *State) Token() (*oauth2.Token, error) {
	s.tokenMu.Lock()
	defer s.tokenMu.Unlock()
	return s.token, nil
}

func (s *State) Client() *github.Client {
	s.tokenMu.Lock()
	defer s.tokenMu.Unlock()

	if s.client == nil {
		var httpClient *http.Client
		if s.token != nil {
			httpClient = oauth2.NewClient(s.ctx, s)
		}
		s.client = github.NewClient(httpClient)
	}
	return s.client
}

func (s *State) Logger() (*gomol.Base, error) {
	if !s.logger.IsInitialized() {
		// err may not be nil if we add more loggers
		if err := s.logger.InitLoggers(); err != nil {
			return nil, err
		}
	}
	return s.logger, nil
}

func (s *State) ShutdownLoggers() error {
	if s == nil {
		return nil
	}

	if s.logger.IsInitialized() {
		return s.logger.ShutdownLoggers()
	}

	return nil
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

func newLogger(out io.Writer) (*gomol.Base, error) {
	consoleConfig := gomolconsole.ConsoleLoggerConfig{
		Colorize: true,
		Writer:   out,
	}

	// err is always nil
	consoleLogger, _ := gomolconsole.NewConsoleLogger(&consoleConfig)
	tpl, _ := gomol.NewTemplate(fullLogTemplate)
	_ = consoleLogger.SetTemplate(tpl)

	logger := gomol.NewBase(
		func(b *gomol.Base) {
			b.SetConfig(
				&gomol.Config{
					FilenameAttr:   "filename",
					LineNumberAttr: "lineno",
					SequenceAttr:   "seq",
					MaxQueueSize:   10000,
				},
			)
		},
	)

	// err is always nil because neither base nor console are initialized
	if err := logger.AddLogger(consoleLogger); err != nil {
		return nil, err
	}

	return logger, nil
}

func usage(flagset *flag.FlagSet) func() {
	return func() {
		fmt.Fprintf(flagset.Output(), "Usage: %s [flags] <dump | list> [template...]\n", flagset.Name())
		flagset.PrintDefaults()
	}
}
