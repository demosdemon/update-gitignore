package app

import (
	"context"
	"fmt"
	"sync"

	"github.com/aphistic/gomol"
)

// Tree yields the gitignore template files from the GitHub repo.
func (s *State) Tree(ctx context.Context) <-chan *Template {
	branch := s.GetDefaultBranch(ctx)
	commit := s.GetBranchHead(ctx, branch)
	return s.getTree(ctx, commit)
}

// GetDefaultBranch returns the default branch for the selected GitHub repo.
func (s *State) GetDefaultBranch(ctx context.Context) string {
	repo, _, err := s.Client().Repositories.Get(ctx, s.Owner, s.Repo)
	if err != nil {
		gomol.Fatalf("Error fetching repo %s/%s.", s.Owner, s.Repo)
		// panic(err)
		return ""
	}

	rv := repo.GetDefaultBranch()
	gomol.Debugf("default branch = %s", rv)
	return rv
}

// GetBranchHead returns the SHA of the
func (s *State) GetBranchHead(ctx context.Context, branchName string) string {
	branch, _, err := s.Client().Repositories.GetBranch(ctx, s.Owner, s.Repo, branchName)
	if err != nil {
		gomol.Fatalf("Error fetching branch %s for repo %s/%s.", branchName, s.Owner, s.Repo)
		// panic(err)
		return "" // TODO: return error
	}

	commit := branch.Commit
	sha := commit.SHA
	gomol.Debugf("head commit = %s", *sha)
	return *sha
}

func (s *State) getTree(ctx context.Context, sha string) <-chan *Template {
	out := make(chan *Template, 5)

	go func() {
		defer close(out)
		wg := new(sync.WaitGroup)

		tree, _, err := s.Client().Git.GetTree(ctx, s.Owner, s.Repo, sha, false)
		if err != nil {
			gomol.Fatalf("Error fetching tree %s", sha)
			// panic(err)
			return // TODO: return error
		}

		for _, entry := range tree.Entries {
			switch Type := entry.GetType(); Type {
			case "blob":
				gitignore := NewTemplate(entry)
				if gitignore != nil {
					select {
					case out <- gitignore:
					case <-ctx.Done():
						// panic(ctx.Err())
						return // TODO: return error
					}
				}
			case "tree":
				wg.Add(1)
				entry := entry // capture loop variable for the closure
				go func() {
					ch := s.getTree(ctx, entry.GetSHA())
					for v := range ch {
						v.Path = fmt.Sprintf("%s/%s", entry.GetPath(), v.Path)
						v.Tags = append(v.Tags, entry.GetPath())
						// don't need to test for ctx.Done because it's always caught above
						out <- v
					}
					wg.Done()
				}()
			default:
				gomol.Warningf("Unknown tree entry type %s %#v", Type, entry)
			}
		}
		wg.Wait()
	}()

	return out
}
