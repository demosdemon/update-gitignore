package app_test

import (
	"bufio"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/google/go-github/v24/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/demosdemon/update-gitignore/app"
)

type replay string

func (r replay) Validate() error {
	st, err := os.Stat(r.Root())
	if err != nil {
		return err
	}

	if st.IsDir() {
		return nil
	}

	return os.ErrInvalid
}

func (r replay) Root() string {
	return string(r)
}

func (r replay) RoundTrip(req *http.Request) (*http.Response, error) {
	defer func() {
		if req.Body != nil {
			req.Body.Close()
		}
	}()

	time.Sleep(time.Microsecond)
	select {
	case <-req.Context().Done():
		return nil, req.Context().Err()
	default:
	}

	if err := r.Validate(); err != nil {
		return nil, err
	}

	if req.Method != http.MethodGet {
		return nil, http.ErrNotSupported
	}

	url := req.URL
	p := path.Join(r.Root(), url.Hostname(), url.Path)

	st, err := os.Stat(p)
	if err != nil {
		return nil, err
	}
	if st.IsDir() {
		p = path.Join(p, "_index")
	}

	fp, err := os.Open(p)
	if err != nil {
		return nil, err
	}

	reader := bufio.NewReader(fp)
	return http.ReadResponse(reader, req)
}

func newReplay(key string) http.RoundTripper {
	root, _ := filepath.Abs("./testdata")
	return replay(filepath.Join(root, key))
}

func newClient(env []string, key string) *app.Client {
	a := newApp(env, "-timeout=0", "test")
	s := app.State{App: a}
	_ = s.ParseArguments()
	c, _ := s.Client()
	c.SetHTTPClient(&http.Client{Transport: newReplay(key)})
	return c
}

func strptr(s string) *string {
	return &s
}

func intts(ts int64) github.Timestamp {
	return github.Timestamp{Time: time.Unix(ts, 0)}
}

func errEquals(tb testing.TB, expected *string, actual error) bool {
	if expected == nil {
		return assert.NoError(tb, actual)
	}

	return assert.EqualError(tb, actual, *expected)
}

func equals(expected, actual interface{}) bool {
	expectedValue := reflect.ValueOf(expected)
	if expectedValue.Kind() == reflect.Ptr {
		expectedValue = expectedValue.Elem()
	}

	actualValue := reflect.ValueOf(actual)
	if actualValue.Kind() == reflect.Ptr {
		actualValue = actualValue.Elem()
	}

	if !expectedValue.IsValid() {
		return !actualValue.IsValid()
	}

	expectedType := expectedValue.Type()

	n := expectedType.NumField()
	for i := 0; i < n; i++ {
		field := expectedType.Field(i)

		e := expectedValue.FieldByName(field.Name)
		a := actualValue.FieldByName(field.Name)
		if a.Kind() == reflect.Invalid {
			return false
		}
		if a.Kind() == reflect.Ptr {
			a = a.Elem()
		}
		if e.Kind() != a.Kind() {
			return false
		}

		switch field.Type.Kind() {
		case reflect.Bool:
			if e.Bool() != a.Bool() {
				return false
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if e.Int() != a.Int() {
				return false
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			if e.Uint() != a.Uint() {
				return false
			}
		case reflect.Float32, reflect.Float64:
			if e.Float() != a.Float() {
				return false
			}
		case reflect.Complex64, reflect.Complex128:
			if e.Complex() != a.Complex() {
				return false
			}
		case reflect.String:
			if e.String() != a.String() {
				return false
			}
		case reflect.Struct:
			if !equals(e.Interface(), a.Interface()) {
				return false
			}
		case reflect.Slice:
			if e.Len() != a.Len() {
				return false
			}
			for i := 0; i < e.Len(); i++ {
				x := e.Index(i)
				y := e.Index(i)
				if x.Kind() == y.Kind() && x.Kind() == reflect.Struct {
					if !equals(x.Interface(), y.Interface()) {
						return false
					}
				} else {
					if !reflect.DeepEqual(x.Interface(), y.Interface()) {
						return false
					}
				}
			}

		default:
			panic(fmt.Errorf("cannot handle kind %v", field.Type.Kind()))
		}
	}

	return true
}

func TestClient_Token(t *testing.T) {
	type fields struct {
		AccessToken string
	}
	cases := []struct {
		name        string
		environment []string
		fields      *fields
		err         *string
	}{
		{
			"no environment",
			nil,
			nil,
			strptr("token not found"),
		},
		{
			"fake key",
			[]string{"GITHUB_TOKEN=fakekey"},
			&fields{"fakekey"},
			nil,
		},
		{
			"empty key",
			[]string{"GITHUB_TOKEN="},
			nil,
			strptr("token not found"),
		},
	}

	t.Parallel()
	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			c := newClient(tt.environment, "invalid")
			require.NotNil(t, c)

			tok, err := c.Token()
			assert.True(t, equals(tt.fields, tok))
			errEquals(t, tt.err, err)
		})
	}
}

func TestClient_GitHubClient(t *testing.T) {
	a := newApp([]string{"GITHUB_TOKEN=faketoken"}, "test")
	s := app.State{App: a}
	err := s.ParseArguments()
	assert.NoError(t, err)
	c, err := s.Client()
	assert.NoError(t, err)

	n := rand.Intn(8) + 2
	ch := make(chan *github.Client, n)

	wg := new(sync.WaitGroup)
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			ch <- c.GitHubClient()
		}()
	}
	wg.Wait()
	close(ch)

	gh := <-ch
	for o := range ch {
		assert.Equal(t, gh, o)
	}

	c.SetHTTPClient(http.DefaultClient)
	assert.NotEqual(t, gh, c.GitHubClient())
}

func TestClient_GetUser(t *testing.T) {
	type fields struct {
		Login   string
		ID      int64
		NodeID  string
		URL     string
		HTMLURL string
	}

	cases := []struct {
		key    string
		fields *fields
		err    *string
	}{
		{
			"valid",
			&fields{
				Login:   "demosdemon",
				ID:      310610,
				NodeID:  "MDQ6VXNlcjMxMDYxMA==",
				URL:     "https://api.github.com/users/demosdemon",
				HTMLURL: "https://github.com/demosdemon",
			},
			nil,
		},
	}

	t.Parallel()
	for _, tt := range cases {
		tt := tt
		t.Run(tt.key, func(t *testing.T) {
			c := newClient(nil, tt.key)
			require.NotNil(t, c)

			user, err := c.GetUser()
			assert.True(t, equals(tt.fields, user))
			errEquals(t, tt.err, err)
		})
	}
}

func TestClient_GetRateLimits(t *testing.T) {
	cases := []struct {
		key string
		rl  *github.RateLimits
		err *string
	}{
		{
			"valid",
			&github.RateLimits{
				Core: &github.Rate{
					Limit:     5000,
					Remaining: 4984,
					Reset:     intts(1553481924),
				},
				Search: &github.Rate{
					Limit:     30,
					Remaining: 30,
					Reset:     intts(1553478384),
				},
			},
			nil,
		},
	}

	t.Parallel()
	for _, tt := range cases {
		tt := tt
		t.Run(tt.key, func(t *testing.T) {
			c := newClient(nil, tt.key)
			require.NotNil(t, c)

			rl, err := c.GetRateLimits()
			assert.EqualValues(t, tt.rl, rl)
			errEquals(t, tt.err, err)
		})
	}
}

func TestClient_GetRepository(t *testing.T) {
	type fields struct {
		ID       int64
		NodeID   string
		Name     string
		FullName string
	}
	cases := []struct {
		key    string
		fields *fields
		err    *string
	}{
		{
			"valid",
			&fields{
				ID:       1062897,
				NodeID:   "MDEwOlJlcG9zaXRvcnkxMDYyODk3",
				Name:     "gitignore",
				FullName: "github/gitignore",
			},
			nil,
		},
	}

	t.Parallel()
	for _, tt := range cases {
		tt := tt
		t.Run(tt.key, func(t *testing.T) {
			c := newClient(nil, tt.key)
			require.NotNil(t, c)

			repo, err := c.GetRepository()
			assert.True(t, equals(tt.fields, repo))
			errEquals(t, tt.err, err)
		})
	}
}

func TestClient_GetBranch(t *testing.T) {
	type commit struct {
		SHA         string
		HTMLURL     string
		URL         string
		CommentsURL string
	}
	type fields struct {
		Name      string
		Commit    commit
		Protected bool
	}
	cases := []struct {
		key    string
		fields *fields
		err    *string
	}{
		{
			"valid",
			&fields{
				Name: "master",
				Commit: commit{
					"56e3f5a7b2a67413a1d3e33fceb8100898015a2e",
					"https://github.com/github/gitignore/commit/56e3f5a7b2a67413a1d3e33fceb8100898015a2e",
					"https://api.github.com/repos/github/gitignore/commits/56e3f5a7b2a67413a1d3e33fceb8100898015a2e",
					"https://api.github.com/repos/github/gitignore/commits/56e3f5a7b2a67413a1d3e33fceb8100898015a2e/comments",
				},
				Protected: false,
			},
			nil,
		},
	}

	t.Parallel()
	for _, tt := range cases {
		tt := tt
		t.Run(tt.key, func(t *testing.T) {
			c := newClient(nil, tt.key)
			require.NotNil(t, c)

			branch, err := c.GetBranch("master")
			assert.True(t, equals(tt.fields, branch))
			errEquals(t, tt.err, err)
		})
	}
}

func TestClient_GetTree(t *testing.T) {
	type TreeEntry struct {
		SHA  string
		Path string
		Mode string
		Type string
		Size int
		URL  string
	}
	type fields struct {
		SHA       string
		Entries   []TreeEntry
		Truncated bool
	}

	cases := []struct {
		key    string
		sha    string
		fields *fields
		err    *string
	}{
		{
			"valid",
			"45f58ef9211cc06f3ef86585c7ecb1b3d52fd4f9",
			nil,
			strptr("Get https://api.github.com/repos/github/gitignore/git/trees/45f58ef9211cc06f3ef86585c7ecb1b3d52fd4f9: unexpected EOF"),
		},
		{
			"valid",
			"c393f60c1f79784dc0660002fc15fc96a64103a7",
			&fields{
				SHA: "c393f60c1f79784dc0660002fc15fc96a64103a7",
				Entries: []TreeEntry{
					{
						Path: "Snap.gitignore",
						Mode: "100644",
						Type: "blob",
						SHA:  "ea38c6dd427cf29cf2635da44d3b4b314c4397ad",
						Size: 363,
						URL:  "https://api.github.com/repos/github/gitignore/git/blobs/ea38c6dd427cf29cf2635da44d3b4b314c4397ad",
					},
				},
				Truncated: false,
			},
			nil,
		},
	}

	t.Parallel()
	for _, tt := range cases {
		tt := tt
		t.Run(tt.key+" "+tt.sha, func(t *testing.T) {
			c := newClient(nil, tt.key)
			require.NotNil(t, c)

			tree, err := c.GetTree(tt.sha)
			assert.True(t, equals(tt.fields, tree))
			errEquals(t, tt.err, err)
		})
	}
}
