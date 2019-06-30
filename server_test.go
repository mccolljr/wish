package wish_test

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/mccolljr/wish"
)

type M = map[string]string

type ServerContext struct {
	wish.ContextImpl
	wish.UtilityFuncs
}

func (c *ServerContext) HandleRoot(w http.ResponseWriter, r *http.Request) {
	c.Respond(w, "text/plain", 200, []byte("is root"))
}

func (c *ServerContext) GetJSON(w http.ResponseWriter, r *http.Request) {
	c.JSON(w, 200, M{"a": "1"})
}

func (c *ServerContext) GetError(w http.ResponseWriter, r *http.Request) {
	c.Error(w, 405)
}

func (c *ServerContext) MountWeb() http.Handler {
	return http.FileServer(http.Dir("testdata/server_files"))
}

func TestServer(t *testing.T) {
	s, err := wish.Bootstrap(func() (wish.Context, error) {
		return &ServerContext{}, nil
	}, wish.DoLog, func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Golang-Test", "1")
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Log("routes:\n" + strings.Join(s.Routes(), "\n"))

	go func() {
		t.Fatal(http.ListenAndServe(":12345", s))
	}()
	time.Sleep(time.Millisecond * 200)

	for i, c := range []struct {
		method      string
		path        string
		wantStatus  int
		wantContent string
		wantHeaders map[string]string
	}{
		{"GET", "/", 200, "is root", nil},
		{"PUT", "/", 200, "is root", nil},
		{"POST", "/", 200, "is root", nil},
		{"PATCH", "/", 200, "is root", nil},
		{"DELETE", "/", 200, "is root", nil},
		{"CONNECT", "/", 200, "is root", nil},
		{"TRACE", "/", 200, "is root", nil},
		{"OPTIONS", "/", 200, "is root", nil},
		{"HEAD", "/", 200, "", nil},

		{"GET", "/json", 200, `{"a":"1"}`, nil},
		{"GET", "/error", 405, "Method Not Allowed", nil},
		{"GET", "/web/a.txt", 200, "is a.txt", nil},
		{"GET", "/web/b.txt", 200, "is b.txt", nil},

		{"GET", "/other", 404, "404 page not found\n", nil},
		{"GET", "/web/other.txt", 404, "404 page not found\n", nil},
	} {
		r, err := http.NewRequest(c.method, "http://localhost:12345"+c.path, nil)
		if err != nil {
			t.Errorf("case %d: error creating request: %v", i, err)
			continue
		}

		resp, err := http.DefaultClient.Do(r)
		if err != nil {
			t.Errorf("case %d: error sending request: %v", i, err)
			continue
		}

		defer resp.Body.Close()

		if resp.StatusCode != c.wantStatus {
			t.Errorf("case %d: wanted status=%d, got status=%d", i, c.wantStatus, resp.StatusCode)
		}

		if content, err := ioutil.ReadAll(resp.Body); err != nil {
			t.Errorf("case %d: error reading body: %v", i, err)
		} else {
			if gotContent := string(content); gotContent != c.wantContent {
				t.Errorf("case %d: wanted content=%q, got content=%q", i, c.wantContent, gotContent)
			}
		}

		for name, wantHeader := range c.wantHeaders {
			gotHeader := resp.Header.Get(name)
			if wantHeader != gotHeader {
				t.Errorf("case %d: wanted header %q=%q, got header %q=%q", i, name, wantHeader, name, gotHeader)
			}
		}
	}
}
