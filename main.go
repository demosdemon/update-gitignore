package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/gosuri/uitable"

	"github.com/demosdemon/update-gitignore/app"
)

var args = os.Args[1:]
var stdout io.Writer = os.Stdout
var stderr io.Writer = os.Stderr

type sysexit func(code int)

var exit sysexit = os.Exit

func main() {
	defer app.PanicOnError(app.InitLogging())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state, err := app.NewState(ctx, args, stderr)
	if err != nil {
		fmt.Fprintf(stderr, "error initializing state: %v\n", err)
		exit(2)
	}

	tree := state.Tree(ctx)

	switch {
	case state.Dump:
	case state.List:
		list(tree, state.Templates...)
	}
}

func list(ch <-chan *app.Template, templates ...string) {
	table := uitable.New()
	f := filter(templates)
	table.AddRow("NAME", "SIZE", "TAGS", "SHA")
	for tpl := range ch {
		if f(tpl) {
			table.AddRow(tpl.Name, tpl.Size, comma(tpl.Tags), tpl.SHA)
		}
	}
	fmt.Fprintln(stdout, table)
}

func comma(v []string) string {
	return strings.Join(v, ", ")
}

func filter(v []string) func(*app.Template) bool {
	pattern := fmt.Sprintf("(?i)(%s)", strings.Join(apply(regexp.QuoteMeta, v), "|"))
	re := regexp.MustCompile(pattern)
	return func(tpl *app.Template) bool {
		if re.Match([]byte(tpl.Name)) {
			return true
		}
		for _, x := range tpl.Tags {
			if re.Match([]byte(x)) {
				return true
			}
		}
		return false
	}
}

func apply(f func(string) string, v []string) []string {
	wg := new(sync.WaitGroup)
	wg.Add(len(v))

	ch := make(chan string, len(v))
	for _, x := range v {
		go func(y string) {
			defer wg.Done()
			ch <- f(y)
		}(x)
	}

	wg.Wait()
	close(ch)

	rv := make([]string, 0, cap(ch))
	for x := range ch {
		rv = append(rv, x)
	}

	return rv
}
