package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/demosdemon/update-gitignore/app"
)

var args = os.Args[1:]

func main() {
	defer app.PanicOnError(app.InitLogging())

	state := app.NewState(args)

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
