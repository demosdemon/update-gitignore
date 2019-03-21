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
func InitLogging() func() {
	consoleCfg := gc.NewConsoleLoggerConfig()
	consoleCfg.Writer = os.Stderr

	if consoleLogger, err := gc.NewConsoleLogger(consoleCfg); err != nil { // nocover
		panic(err)
	} else {
		tpl := gc.NewTemplateFull()
		if tpl != nil {
			if err := consoleLogger.SetTemplate(tpl); err != nil { // nocover
				panic(err)
			}
		}
		gomol.AddLogger(consoleLogger)
	}

	if err := gomol.InitLoggers(); err != nil { // nocover
		panic(err)
	}

	return func() {
		gomol.ShutdownLoggers()
	}
}
