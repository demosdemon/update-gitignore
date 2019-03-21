package gitignore

import (
	"path"
	"strings"

	"github.com/google/go-github/v24/github"
)

const (
	// Suffix is the file suffix to grok.
	Suffix = ".gitignore"
)

// Template represents a gitignore template from the repository
type Template struct {
	Name string
	Size uint64
	Path string
	Tags []string
	SHA  string
}

// New builds a new Template struct from a GitHub TreeEntry. Returns nil if
// the entry doesn't point to a Gitignore template.
func New(entry github.TreeEntry) *Template {
	Size := uint64(entry.GetSize())
	Path := entry.GetPath()
	SHA := entry.GetSHA()

	if strings.HasSuffix(Path, Suffix) {
		_, basename := path.Split(Path)
		Name := strings.TrimSuffix(basename, Suffix)
		if Name == "" {
			return nil
		}

		return &Template{
			Name,
			Size,
			Path,
			nil,
			SHA,
		}
	}

	return nil
}
