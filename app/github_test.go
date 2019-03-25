package app

import (
	"bufio"
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"testing"

	"github.com/google/go-github/v24/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type roundtripFunc func(*http.Request) (*http.Response, error)

func (f roundtripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type buffer bytes.Buffer

func (b *buffer) Read(p []byte) (n int, err error) {
	return (*bytes.Buffer)(b).Read(p)
}

func (b *buffer) Close() error {
	return nil
}

func testdataRoundTrip(req *http.Request) (*http.Response, error) {
	notFound := &http.Response{
		Status:     "404 NOT FOUND",
		StatusCode: 404,
		Proto:      "HTTP/1.0",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Request:    req,
		Body:       &buffer{},
	}

	if req.Method != http.MethodGet {
		return notFound, nil
	}

	requestPath := path.Join(".", "testdata", req.URL.Hostname(), req.URL.Path)
	pathInfo, err := os.Stat(requestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return notFound, nil
		}

		return nil, err
	}

	if pathInfo.IsDir() {
		requestPath = path.Join(requestPath, "_index")
	}

	data, err := ioutil.ReadFile(requestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return notFound, nil
		}

		return nil, err
	}

	buf := bytes.NewBuffer(data)
	reader := bufio.NewReader(buf)
	resp, err := http.ReadResponse(reader, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

var testdataTransport = roundtripFunc(testdataRoundTrip)
var httpClient = http.Client{Transport: testdataTransport}
var githubClient = github.NewClient(&httpClient)

func TestGithubClient(t *testing.T) {
	user, res, err := githubClient.Users.Get(context.Background(), "")
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, "GitHub.com", res.Header.Get("Server"))
	assert.Equal(t, "Sun, 24 Mar 2019 21:21:33 GMT", res.Header.Get("Date"))
	assert.Equal(t, `"b2f5a7c1c770b449c70f83bccb5f3495"`, res.Header.Get("ETag"))
	assert.Equal(t, "Wed, 06 Mar 2019 20:51:13 GMT", res.Header.Get("Last-Modified"))
	assert.Equal(t, "E7D7:24FD:CD092B:203716C:5C97F4DD", res.Header.Get("X-GitHub-Request-Id"))
	require.NotNil(t, user)
	assert.Equal(t, "demosdemon", user.GetLogin())
	assert.Equal(t, int64(310610), user.GetID())
	assert.Equal(t, "MDQ6VXNlcjMxMDYxMA==", user.GetNodeID())
}

func TestGithubClientIndex(t *testing.T) {
	repo, res, err := githubClient.Repositories.Get(context.Background(), "github", "gitignore")
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, "GitHub.com", res.Header.Get("Server"))
	assert.Equal(t, "Sun, 24 Mar 2019 21:26:07 GMT", res.Header.Get("Date"))
	assert.Equal(t, `"a765259a0bf4988ea92c7d4d3603c3ad"`, res.Header.Get("ETag"))
	assert.Equal(t, "Sun, 24 Mar 2019 21:23:49 GMT", res.Header.Get("Last-Modified"))
	assert.Equal(t, "E7EC:3C91:156C708:2FC4FDB:5C97F5EF", res.Header.Get("X-GitHub-Request-Id"))
	require.NotNil(t, repo)
	assert.Equal(t, int64(1062897), repo.GetID())
	assert.Equal(t, "MDEwOlJlcG9zaXRvcnkxMDYyODk3", repo.GetNodeID())
	assert.Equal(t, "gitignore", repo.GetName())
}

func TestGithubDefaultBranch(t *testing.T) {
	state, err := New(context.Background(), []string{"list"}, nil, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, state)
	state.client = githubClient

	defaultBranch := state.GetDefaultBranch(context.Background())
	assert.Equal(t, "master", defaultBranch)

	state.Repo = "anything"
	defaultBranch = state.GetDefaultBranch(context.Background())
	assert.Zero(t, defaultBranch)
}

func TestGithubBranchHead(t *testing.T) {
	state, err := New(context.Background(), []string{"list"}, nil, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, state)
	state.client = githubClient

	commit := state.GetBranchHead(context.Background(), "master")
	assert.Equal(t, "56e3f5a7b2a67413a1d3e33fceb8100898015a2e", commit)

	commit = state.GetBranchHead(context.Background(), "anything")
	assert.Zero(t, commit)
}

func TestGithubTree(t *testing.T) {
	state, err := New(context.Background(), []string{"list"}, nil, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, state)
	state.client = githubClient

	ch := state.Tree(context.Background())

	res := make([]*Template, 0, 223)
	for x := range ch {
		res = append(res, x)
	}

	assert.Len(t, res, 223)
}
