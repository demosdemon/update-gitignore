package app

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/aphistic/gomol"
)

var (
	debug = flag.Bool("debug", false, "Print debug statements to STDERR.")
	dump  = flag.Bool("dump", false, "Dump the specified templates to STDOUT.")
	list  = flag.Bool("list", false, "List the available templates. If any, arguments are used to filter the results.")
	repo  = flag.String("repo", "github/gitignore", "The template repo to use.")
)

// The State of the application.
type State struct {
	// command line arguments
	Dump      bool
	List      bool
	Templates []string

	// calculated from cmdline -repo
	Owner string
	Repo  string
}

// NewState build a new application state object
func NewState(arguments []string) *State {
	*debug = false
	*dump = false
	*list = false

	flag.Usage = usage
	// use a var so args can be mocked in tests
	_ = flag.CommandLine.Parse(arguments)

	if *debug {
		gomol.SetLogLevel(gomol.LevelDebug)
	} else {
		gomol.SetLogLevel(gomol.LevelInfo)
	}

	slice := strings.SplitN(*repo, "/", 2)
	if len(slice) != 2 {
		gomol.Dief(2, "Invalid repo (%v). Expected a name in the form `<owner>/<repo>`.", slice)
	}

	state := &State{
		Dump:      *dump,
		List:      *list,
		Templates: flag.Args(),
		Owner:     slice[0],
		Repo:      slice[1],
	}

	// command line options
	gomol.Debugf("Dump       = %t", state.Dump)
	gomol.Debugf("List       = %t", state.List)
	gomol.Debugf("Owner      = %s", state.Owner)
	gomol.Debugf("Repo       = %s", state.Repo)
	gomol.Debugf("Templates  = %v", state.Templates)
	// XXX: add os.Environ()?
	gomol.Debugf("executable = %s", stringOrError(os.Executable))
	gomol.Debugf("euid       = %d", os.Geteuid())
	gomol.Debugf("euid       = %d", os.Geteuid())
	gomol.Debugf("egid       = %d", os.Getegid())
	gomol.Debugf("uid        = %d", os.Getuid())
	gomol.Debugf("gid        = %d", os.Getgid())
	// XXX: add os.Groups()
	gomol.Debugf("pid        = %d", os.Getpid())
	gomol.Debugf("ppid       = %d", os.Getppid())
	gomol.Debugf("cwd        = %s", stringOrError(os.Getwd))
	gomol.Debugf("hostname   = %s", stringOrError(os.Hostname))

	if state.Dump && state.List {
		gomol.Die(1, "-dump and -list are mutually exclusive.")
	}

	if !state.Dump && !state.List {
		gomol.Die(1, "one of -dump or -list is required.")
	}

	if state.Dump && len(state.Templates) == 0 {
		gomol.Die(1, "Must provide at least one template with -dump.")
	}

	return state
}

// Usage prints command line usage information.
func usage() {
	fmt.Printf("usage: %s [flags] [template ...]\n", os.Args[0])
	flag.PrintDefaults()
}
