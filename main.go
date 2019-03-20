package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/demosdemon/update-gitignore/app"
)

func main() {
	shutdown := app.InitLogging()
	defer shutdown()

	state := app.NewState()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()
	tree := state.Tree(ctx)

	switch {
	case state.Dump:
	case state.List:
		for v := range tree {
			fmt.Printf("%+v\n", v)
		}
	default:
		panic(errors.New("how did we get here"))
	}
}
