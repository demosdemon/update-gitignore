package app

import (
	"testing"

	"github.com/google/go-github/v24/github"
	"github.com/stretchr/testify/assert"
)

func TestNewTemplate(t *testing.T) {
	defer PanicOnError(InitLogging())

	sha := "0000000000000000000000000000000000000000"
	size := 42

	t.Run("ValidEntry", func(t *testing.T) {
		path := "Python.gitignore"
		entry := github.TreeEntry{
			SHA:  &sha,
			Path: &path,
			Size: &size,
		}
		tpl := NewTemplate(entry)
		assert.NotNil(t, tpl)
		assert.EqualValues(t, "Python", tpl.Name)
		assert.EqualValues(t, size, tpl.Size)
		assert.EqualValues(t, path, tpl.Path)
		assert.Nil(t, tpl.Tags)
		assert.EqualValues(t, sha, tpl.SHA)
	})

	t.Run("BareGitignore", func(t *testing.T) {
		path := ".gitignore"
		entry := github.TreeEntry{
			SHA:  &sha,
			Path: &path,
			Size: &size,
		}
		tpl := NewTemplate(entry)
		assert.Nil(t, tpl)
	})

	t.Run("InvalidPath", func(t *testing.T) {
		path := ".travis.yml"
		entry := github.TreeEntry{
			SHA:  &sha,
			Path: &path,
			Size: &size,
		}
		tpl := NewTemplate(entry)
		assert.Nil(t, tpl)
	})
}
