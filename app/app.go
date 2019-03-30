package app

import (
	"context"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/aphistic/gomol"
	gomolconsole "github.com/aphistic/gomol-console"
)

const (
	logTemplate = `{{.Template.Format "2006-01-02 15:04:05.000"}} [{{color}}{{ucase .LevelName}}{{reset}}] {{.Message}}`

	fullLogTemplate = logTemplate + `{{if .Attrs}} {{json .Attrs}}{{end}}`
)

type App struct {
	Arguments   []string
	Environment []string
	Context     context.Context
	Stdin       io.Reader
	Stdout      io.Writer
	Stderr      io.Writer
	Exit        func(int)

	loggerMu sync.Mutex
	logger   *gomol.Base

	errchMu sync.Mutex
	errch   chan error
}

func New() *App {
	return &App{
		Arguments:   os.Args[1:],
		Environment: os.Environ(),
		Context:     context.Background(),
		Stdin:       os.Stdin,
		Stdout:      os.Stdout,
		Stderr:      os.Stderr,
		Exit:        os.Exit,
	}
}

func (a *App) Logger() *gomol.Base {
	a.loggerMu.Lock()
	defer a.loggerMu.Unlock()

	if a.logger == nil {
		consoleConfig := gomolconsole.ConsoleLoggerConfig{
			Colorize: true,
			Writer:   a.Stderr,
		}

		// err is always nil
		consoleLogger, _ := gomolconsole.NewConsoleLogger(&consoleConfig)

		// err is always nil because the template is not dynamic and I tested it at least once
		tpl, _ := gomol.NewTemplate(fullLogTemplate)

		// err is always nil if the template is non-nil
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

		// err is always nil since we're not reusing objects
		_ = logger.AddLogger(consoleLogger)

		a.logger = logger
	}

	return a.logger
}

func (a *App) ensureErrorChannel() {
	a.errchMu.Lock()
	defer a.errchMu.Unlock()

	if a.errch == nil {
		a.errch = make(chan error, 1)
	}
}

func (a *App) Errors() <-chan error {
	a.ensureErrorChannel()
	return a.errch
}

func (a *App) HandleError(err error) {
	a.ensureErrorChannel()
	defer close(a.errch)
	select {
	case a.errch <- err:
	case <-a.Context.Done():
	}
}

func (a *App) LookupEnv(key string) (string, bool) {
	ch := make(chan string)

	wg := sync.WaitGroup{}
	wg.Add(len(a.Environment))

	go func() {
		for _, line := range a.Environment {
			line := line
			go func() {
				defer wg.Done()
				select {
				case <-a.Context.Done():
					return
				default:
					slice := strings.SplitN(line, "=", 2)
					if len(slice) == 2 && slice[0] == key {
						ch <- slice[1]
					}
				}
			}()
		}
	}()

	go func() {
		defer close(ch)
		wg.Wait()
	}()

	v, ok := <-ch
	return v, ok
}
