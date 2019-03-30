package main

import (
	"github.com/demosdemon/update-gitignore/v0/app"
)

var instance = app.New()

func main() {
	defer instance.Logger().ShutdownLoggers()
	done := make(chan struct{})

	go func() {
		for err := range instance.Errors() {
			instance.Logger().Error(err.Error())
		}

		done <- struct{}{}
	}()

	state := app.State{App: instance}
	err := state.ParseArguments()
	if err != nil {
		state.HandleError(err)
		<-done
		instance.Exit(2)
	}

	cmd, err := state.Command()
	if err != nil {
		state.HandleError(err)
		<-done
		instance.Exit(2)
	}

	instance.Exit(int(cmd.Run()))
}
