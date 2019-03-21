package app

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/aphistic/gomol"
	"github.com/google/go-github/v24/github"
	"golang.org/x/oauth2"

	"github.com/demosdemon/update-gitignore/gitignore"
)

// Client fetches and caches a GitHub client. If the environment variable
// GITHUB_TOKEN is found, uses it to authenticate GitHub API requests. An API
// token is not required; however, advised due to API rate-limiting.
func (s *State) Client(ctx context.Context) *github.Client {
	if s == nil {
		return nil
	}

	if token, found := os.LookupEnv("GITHUB_TOKEN"); found {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		tc := oauth2.NewClient(ctx, ts)
		return github.NewClient(tc)
	}
	return github.NewClient(nil)
}

// Tree yields the gitignore template files from the GitHub repo.
func (s *State) Tree(ctx context.Context) <-chan *gitignore.Template {
	branch := s.GetDefaultBranch(ctx)
	commit := s.GetBranchHead(ctx, branch)
	return s.getTree(ctx, commit)
}

// GetDefaultBranch returns the default branch for the selected GitHub repo.
func (s *State) GetDefaultBranch(ctx context.Context) string {
	repo, _, err := s.Client(ctx).Repositories.Get(ctx, s.Owner, s.Repo)
	if err != nil {
		gomol.Fatalf("Error fetching repo %s/%s.", s.Owner, s.Repo)
		panicUnlessCanceled(ctx, err)
		return ""
	}

	rv := repo.GetDefaultBranch()
	gomol.Debugf("default branch = %s", rv)
	return rv
}

// GetBranchHead returns the SHA of the
func (s *State) GetBranchHead(ctx context.Context, branchName string) string {
	branch, _, err := s.Client(ctx).Repositories.GetBranch(ctx, s.Owner, s.Repo, branchName)
	if err != nil {
		gomol.Fatalf("Error fetching branch %s for repo %s/%s.", branchName, s.Owner, s.Repo)
		panicUnlessCanceled(ctx, err)
		return ""
	}

	commit := branch.Commit
	if commit == nil {
		panic(fmt.Errorf("got nil for branch.Commit: %#v", branch))
	}

	sha := commit.SHA
	if sha == nil {
		panic(fmt.Errorf("got nil for branch.Commit.SHA: %#v", branch))
	}

	rv := *sha
	gomol.Debugf("head commit = %s", rv)
	return rv
}

func (s *State) getTree(ctx context.Context, sha string) <-chan *gitignore.Template {
	out := make(chan *gitignore.Template, 5)

	go func() {
		defer close(out)
		wg := new(sync.WaitGroup)

		tree, _, err := s.Client(ctx).Git.GetTree(ctx, s.Owner, s.Repo, sha, false)
		if err != nil {
			gomol.Fatalf("Error fetching tree %s", sha)
			panicUnlessCanceled(ctx, err)
			return
		}

		for _, entry := range tree.Entries {
			switch Type := entry.GetType(); Type {
			case "blob":
				gitignore := gitignore.New(entry)
				if gitignore != nil {
					select {
					case out <- gitignore:
					case <-ctx.Done():
						gomol.Warning("getTree loop canceled")
						return
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

func panicUnlessCanceled(ctx context.Context, err error) {
	select {
	case <-ctx.Done():
		gomol.Warning(err.Error())
		return
	default:
		panic(err)
	}
}
