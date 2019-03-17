package main

// Gitignore represents a gitignore template from the repository
type Gitignore struct {
	Name string
	Size uint64
	Path string
	Tags []string
	SHA  string

	state *State
}
