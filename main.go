package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/demosdemon/update-gitignore/app"
)

type sysexit func(code int)

var (
	args             = os.Args[1:]
	stdin  io.Reader = os.Stdin
	stdout io.Writer = os.Stdout
	stderr io.Writer = os.Stderr
	exit   sysexit   = os.Exit
)

func main() {
	var exitCode app.ExitStatus
	defer func() {
		exit(int(exitCode))
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	state, err := app.New(ctx, args, stdin, stdout, stderr)
	defer func() {
		_ = state.ShutdownLoggers()
	}()

	if err != nil {
		fmt.Fprintf(stderr, "error initializing state: %v\n", err)
		exitCode = 2
		return
	}

	cmd, err := state.Command()
	if err != nil {
		fmt.Fprintf(stderr, "error initializing command: %v\n", err)
		exitCode = 3
		return
	}

	exitCode = cmd.Run()
}
