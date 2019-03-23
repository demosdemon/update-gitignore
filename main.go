package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/demosdemon/update-gitignore/app"
	"github.com/gosuri/uitable"
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
		list(tree, state.Templates...)
	default:
		panic(errors.New("how did we get here")) // nocover
	}
}

func list(ch <-chan *app.Template, templates ...string) {
	table := uitable.New()
	f := filter(templates)
	table.AddRow("NAME", "SIZE", "TAGS", "SHA")
	for tpl := range ch {
		if f(tpl) {
			table.AddRow(tpl.Name, tpl.Size, comma(tpl.Tags...), tpl.SHA)
		}
	}
	fmt.Println(table)
}

func comma(v ...string) string {
	return strings.Join(v, ", ")
}

func filter(v []string) func(*app.Template) bool {
	pattern := fmt.Sprintf("(?i)(%s)", strings.Join(apply(regexp.QuoteMeta, v...), "|"))
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

func apply(f func(string) string, v ...string) []string {
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
