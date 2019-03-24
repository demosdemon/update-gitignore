package app

import (
	"testing"

	"github.com/google/go-github/v24/github"
	"github.com/stretchr/testify/assert"
)

var (
	sha  = "0000000000000000000000000000000000000000"
	size = 42
)

func TestNewTemplateWithValidEntry(t *testing.T) {
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
}

func TestNewTemplateWithBareGitignore(t *testing.T) {
	path := ".gitignore"
	entry := github.TreeEntry{
		SHA:  &sha,
		Path: &path,
		Size: &size,
	}

	tpl := NewTemplate(entry)
	assert.Nil(t, tpl)
}

func TestNewTemplateWithInvalidPath(t *testing.T) {
	path := ".travis.yml"
	entry := github.TreeEntry{
		SHA:  &sha,
		Path: &path,
		Size: &size,
	}

	tpl := NewTemplate(entry)
	assert.Nil(t, tpl)
}
