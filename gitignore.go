package main

import (
	"path"
	"strings"

	"github.com/google/go-github/v24/github"
)

// Gitignore represents a gitignore template from the repository
type Gitignore struct {
	Name string
	Size uint64
	Path string
	Tags []string
	SHA  string

	state *State
}

func (s *State) makeGitignore(entry github.TreeEntry) *Gitignore {
	Size := uint64(entry.GetSize())
	Path := entry.GetPath()
	SHA := entry.GetSHA()

	if strings.HasSuffix(Path, Suffix) {
		dir, basename := path.Split(Path)
		Name := strings.TrimSuffix(basename, Suffix)
		if Name == "" {
			return nil
		}

		dir = strings.TrimRight(dir, "/")
		Tags := strings.Split(dir, "/")

		return &Gitignore{
			Name,
			Size,
			Path,
			Tags,
			SHA,
			s,
		}
	}

	return nil
}
