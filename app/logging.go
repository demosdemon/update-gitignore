package app

import (
	"os"

	"github.com/aphistic/gomol"
	gc "gopkg.in/aphistic/gomol-console.v0"
)

func stringOrError(call func() (string, error)) string {
	if rv, err := call(); err == nil {
		return rv
	} else {
		return err.Error()
	}
}

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
	_ = gomol.Debugf("executable = %s", stringOrError(os.Executable))
	_ = gomol.Debugf("euid = %d", os.Geteuid())
	_ = gomol.Debugf("euid = %d", os.Geteuid())
	_ = gomol.Debugf("egid = %d", os.Getegid())
	_ = gomol.Debugf("uid = %d", os.Getuid())
	_ = gomol.Debugf("gid = %d", os.Getgid())
	// XXX: add os.Groups()
	_ = gomol.Debugf("pid = %d", os.Getpid())
	_ = gomol.Debugf("ppid = %d", os.Getppid())
	_ = gomol.Debugf("cwd = %s", stringOrError(os.Getwd))
	_ = gomol.Debugf("hostname = %s", stringOrError(os.Hostname))

	return func() {
		_ = gomol.ShutdownLoggers()
	}
}
