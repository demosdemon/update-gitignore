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

	if consoleLogger, err := gc.NewConsoleLogger(consoleCfg); err != nil {
		panic(err)
	} else {
		tpl := gc.NewTemplateFull()
		if tpl != nil {
			if err := consoleLogger.SetTemplate(tpl); err != nil {
				panic(err)
			}
		}
		gomol.AddLogger(consoleLogger)
	}

	if err := gomol.InitLoggers(); err != nil {
		panic(err)
	}

	// XXX: add os.Environ()?
	gomol.Debugf("executable = %s", stringOrError(os.Executable))
	gomol.Debugf("euid = %d", os.Geteuid())
	gomol.Debugf("euid = %d", os.Geteuid())
	gomol.Debugf("egid = %d", os.Getegid())
	gomol.Debugf("uid = %d", os.Getuid())
	gomol.Debugf("gid = %d", os.Getgid())
	// XXX: add os.Groups()
	gomol.Debugf("pid = %d", os.Getpid())
	gomol.Debugf("ppid = %d", os.Getppid())
	gomol.Debugf("cwd = %s", stringOrError(os.Getwd))
	gomol.Debugf("hostname = %s", stringOrError(os.Hostname))

	return func() {
		gomol.ShutdownLoggers()
	}
}
