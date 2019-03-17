package main

import (
	"fmt"
)

func main() {
	state := NewState()
	tree := state.Tree()

	switch {
	case state.Dump:
	case state.List:
		for k, v := range tree {
			fmt.Printf("%s = %v\n", k, v)
		}
	default:
		state.Fatal("One of -dump or -list is required.")
	}
}
