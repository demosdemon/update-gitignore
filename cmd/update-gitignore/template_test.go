package main

import (
	"encoding/json"
	"testing"

	"github.com/google/go-github/v24/github"
	"github.com/stretchr/testify/require"
)

func TestNewTemplate(t *testing.T) {
	type testcase struct {
		name      string
		entryJSON string
		expected  *Template
	}

	cases := []testcase{
		{
			"tree",
			`{
				"path": ".github",
				"mode": "040000",
				"type": "tree",
				"sha": "45f58ef9211cc06f3ef86585c7ecb1b3d52fd4f9",
				"url":
				"https://api.github.com/repos/github/gitignore/git/trees/45f58ef9211cc06f3ef86585c7ecb1b3d52fd4f9"
			}`,
			nil,
		},
		{
			"travis.yml",
			`{
				"path": ".travis.yml",
				"mode": "100644",
				"type": "blob",
				"sha": "f362d6fe3228d49e1658e8e66ffbd8ec52ab86c7",
				"size": 103,
				"url":
				"https://api.github.com/repos/github/gitignore/git/blobs/f362d6fe3228d49e1658e8e66ffbd8ec52ab86c7"
			}`,
			nil,
		},
		{
			"actionscript",
			`{
				"path": "Actionscript.gitignore",
				"mode": "100644",
				"sha": "5d947ca8879f8a9072fe485c566204e3c2929e80",
				"size": 350,
				"url":
				"https://api.github.com/repos/github/gitignore/git/blobs/5d947ca8879f8a9072fe485c566204e3c2929e80"
			}`,
			&Template{
				"Actionscript",
				350,
				"Actionscript.gitignore",
				nil,
				"5d947ca8879f8a9072fe485c566204e3c2929e80",
			},
		},
		{
			"gitignore",
			`{
				"path": ".gitignore",
				"mode": "100644",
				"type": "blob",
				"sha": "eba3b7ec8af85c99af1a6c8f375c2ba8a92befc6",
				"size": 3601,
				"url":
				"https://api.github.com/repos/demosdemon/dotfiles/git/blobs/eba3b7ec8af85c99af1a6c8f375c2ba8a92befc6"
			}`,
			nil,
		},
	}

	t.Parallel()
	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var entry github.TreeEntry
			err := json.Unmarshal([]byte(tt.entryJSON), &entry)
			require.NoError(t, err)
			tpl := NewTemplate(entry)
			require.Equal(t, tt.expected, tpl)
		})
	}
}
