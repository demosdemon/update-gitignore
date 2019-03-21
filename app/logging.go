package app

import (
	"os"

	"github.com/aphistic/gomol"
	gc "gopkg.in/aphistic/gomol-console.v0"
)

func stringOrError(call func() (string, error)) string {
	rv, err := call()
	if err == nil {
		return rv
	}
	return err.Error()
}

// InitLogging initializes the gomol logging system for the application. Returns
// a shutdown function that should be called before the app terminates.
func InitLogging() func() error {
	consoleCfg := gc.NewConsoleLoggerConfig()
	consoleCfg.Writer = os.Stderr

	// err is always nil
	consoleLogger, _ := gc.NewConsoleLogger(consoleCfg)
	tpl := gc.NewTemplateFull()
	if err := consoleLogger.SetTemplate(tpl); err != nil {
		panic(err)
	}
	gomol.AddLogger(consoleLogger)

	if err := gomol.InitLoggers(); err != nil {
		panic(err)
	}

	return shutdownLoggers
}

func shutdownLoggers() error {
	return gomol.ShutdownLoggers()
}
