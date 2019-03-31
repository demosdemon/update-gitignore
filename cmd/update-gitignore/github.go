package main

import (
	"errors"
	"net/http"
	"sync"

	"github.com/google/go-github/v24/github"
	"golang.org/x/oauth2"
)

var (
	// ErrTokenNotFound is returned if the client was unable to locate a GITHUB_TOKEN environment variable
	ErrTokenNotFound = errors.New("token not found")
)

type Client struct {
	state      *State
	owner      string
	repo       string
	httpClient *http.Client

	clientMu sync.Mutex
	client   *github.Client
}

func (c *Client) Token() (*oauth2.Token, error) {
	value, _ := c.state.LookupEnv("GITHUB_TOKEN")
	if value == "" {
		return nil, ErrTokenNotFound
	}

	return &oauth2.Token{AccessToken: value}, nil
}

func (c *Client) SetHTTPClient(httpClient *http.Client) {
	if httpClient == nil {
		_, err := c.Token()
		if err == nil {
			httpClient = oauth2.NewClient(c.state.Context, c)
		}
	}

	c.clientMu.Lock()
	c.client = nil
	c.httpClient = httpClient
	c.clientMu.Unlock()
}

func (c *Client) HTTPClient() *http.Client {
	return c.httpClient
}

func (c *Client) GitHubClient() *github.Client {
	c.clientMu.Lock()
	defer c.clientMu.Unlock()
	if c.client == nil {
		c.client = github.NewClient(c.HTTPClient())
	}
	return c.client
}

func (c *Client) GetUser() (*github.User, error) {
	cl := c.GitHubClient()
	ctx, cancel := c.state.deadline()
	defer cancel()
	user, _, err := cl.Users.Get(ctx, "")
	return user, err
}

func (c *Client) GetRateLimits() (*github.RateLimits, error) {
	cl := c.GitHubClient()
	ctx, cancel := c.state.deadline()
	defer cancel()
	rl, _, err := cl.RateLimits(ctx)
	return rl, err
}

func (c *Client) GetRepository() (*github.Repository, error) {
	cl := c.GitHubClient()
	ctx, cancel := c.state.deadline()
	defer cancel()
	repo, _, err := cl.Repositories.Get(ctx, c.owner, c.repo)
	return repo, err
}

func (c *Client) GetBranch(branch string) (*github.Branch, error) {
	cl := c.GitHubClient()
	ctx, cancel := c.state.deadline()
	defer cancel()
	b, _, err := cl.Repositories.GetBranch(ctx, c.owner, c.repo, branch)
	return b, err
}

func (c *Client) GetTree(sha string) (*github.Tree, error) {
	cl := c.GitHubClient()
	ctx, cancel := c.state.deadline()
	defer cancel()
	tree, _, err := cl.Git.GetTree(ctx, c.owner, c.repo, sha, false)
	return tree, err
}

func (c *Client) GetBlob(sha string) (*github.Blob, error) {
	cl := c.GitHubClient()
	ctx, cancel := c.state.deadline()
	defer cancel()
	blob, _, err := cl.Git.GetBlob(ctx, c.owner, c.repo, sha)
	return blob, err
}
