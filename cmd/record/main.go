package main

import (
	"github.com/demosdemon/golang-app-framework/app"
)

var instance = app.New()

func main() {
	defer instance.Logger().ShutdownLoggers()
	instance.Logger().Info("Hello, World!")
}
