package wish

import (
	"errors"
	"net/http"
	"reflect"
	"sort"
	"testing"

	"github.com/go-chi/chi"
)

func TestParseHandler(t *testing.T) {
	for i, c := range []struct {
		name       string
		wantMethod string
		wantPat    string
		wantOk     bool
	}{
		// all http methods, plus special case "HANDLE"
		{"GetRoot", "GET", "/", true},
		{"PutRoot", "PUT", "/", true},
		{"PostRoot", "POST", "/", true},
		{"PatchRoot", "PATCH", "/", true},
		{"DeleteRoot", "DELETE", "/", true},
		{"TraceRoot", "TRACE", "/", true},
		{"ConnectRoot", "CONNECT", "/", true},
		{"OptionsRoot", "OPTIONS", "/", true},
		{"HeadRoot", "HEAD", "/", true},
		{"HandleRoot", "HANDLE", "/", true},

		// complex routes
		{"GetRootByID", "GET", "/{id}", true},
		{"GetOther", "GET", "/other", true},
		{"GetOtherJSON", "GET", "/other/json", true},
		{"GetABBRIncluded", "GET", "/abbr/included", true},
		{"GetOtherByName", "GET", "/other/{name}", true},
		{"GetMultiPartRoute", "GET", "/multi/part/route", true},
		{"GetMultiPartRouteByOnlyOneParamPart", "GET", "/multi/part/route/{onlyoneparampart}", true},

		// bad routes
		{"BadMethodName", "", "", false},
	} {
		gotMethod, gotPat, gotOk := parseHandler(c.name)
		if gotOk != c.wantOk {
			t.Errorf("case %d: expected ok=%t, got ok=%t", i, c.wantOk, gotOk)
		}

		if gotMethod != c.wantMethod {
			t.Errorf("case %d: method: expected %q, got %q", i, c.wantMethod, gotMethod)
		}

		if gotPat != c.wantPat {
			t.Errorf("case %d: pattern: expected %q, got %q", i, c.wantPat, gotPat)
		}
	}
}
func TestParseMount(t *testing.T) {
	for i, c := range []struct {
		name    string
		wantPat string
		wantOk  bool
	}{
		// ok patterns
		{"MountRoot", "/", true},
		{"MountOther", "/other", true},
		{"MountMultiPartRoute", "/multi/part/route", true},
		{"MountWebByParam", "/web/by/param", true},

		// bad patterns
		{"BadPrefix", "", false},
		{"GetRoot", "", false},
	} {
		gotPat, gotOk := parseMount(c.name)
		if gotOk != c.wantOk {
			t.Errorf("case %d: expected ok=%t, got ok=%t", i, c.wantOk, gotOk)
		}

		if gotPat != c.wantPat {
			t.Errorf("case %d: pattern: expected %q, got %q", i, c.wantPat, gotPat)
		}
	}
}

func TestBootstrapInvalidType(t *testing.T) {
	_, err := Bootstrap(func() (Context, error) {
		return struct{ *ContextImpl }{}, nil
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	expected := "bootstrap: expected pointer to a struct, got struct { *genie.ContextImpl }"
	if err.Error() != expected {
		t.Fatalf("expected %q, got %q", expected, err)
	}
}

func TestBootstrapProviderError(t *testing.T) {
	providerErr := errors.New("provider error")
	_, err := Bootstrap(func() (Context, error) {
		return nil, providerErr
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	expected := "bootstrap: " + providerErr.Error()
	if err.Error() != expected {
		t.Fatalf("expected %q, got %q", expected, err)
	}
}

type TestingContext struct {
	ContextImpl
}

func (t *TestingContext) HandleRoot(w http.ResponseWriter, r *http.Request) {}

func (t *TestingContext) MountWeb() http.Handler { return nil }

func (t *TestingContext) UseOther() []MiddlewareFunc {
	return []MiddlewareFunc{
		func(h http.Handler) http.Handler { return h },
	}
}
func (t *TestingContext) GetOtherByID(w http.ResponseWriter, r *http.Request) {}

func TestBootstrap(t *testing.T) {
	s, err := Bootstrap(func() (Context, error) {
		return &TestingContext{}, nil
	})

	if err != nil {
		t.Fatal(err)
	}

	expectedRoutes := []string{"/", "/web/*", "/other/{id}"}

	gotRoutes := s.Routes()
	sort.Slice(gotRoutes, func(a, b int) bool {
		av, bv := gotRoutes[a], gotRoutes[b]
		al, bl := len(av), len(bv)
		if al == bl {
			return av < bv
		}
		return al < bl
	})

	if !reflect.DeepEqual(gotRoutes, expectedRoutes) {
		t.Fatalf("expected routes %v, got %v", expectedRoutes, gotRoutes)
	}

	for i, c := range []struct {
		method, path string
		wantMatch    bool
		wantParams   map[string]string
	}{
		{"GET", "/", true, nil},
		{"PUT", "/", true, nil},
		{"POST", "/", true, nil},
		{"PATCH", "/", true, nil},
		{"DELETE", "/", true, nil},
		{"TRACE", "/", true, nil},
		{"CONNECT", "/", true, nil},
		{"OPTIONS", "/", true, nil},
		{"HEAD", "/", true, nil},

		{"GET", "/web/blah.txt", true, map[string]string{"*": "blah.txt"}},
		{"PUT", "/web/blah.txt", true, map[string]string{"*": "blah.txt"}},
		{"POST", "/web/blah.txt", true, map[string]string{"*": "blah.txt"}},
		{"PATCH", "/web/blah.txt", true, map[string]string{"*": "blah.txt"}},
		{"DELETE", "/web/blah.txt", true, map[string]string{"*": "blah.txt"}},
		{"TRACE", "/web/blah.txt", true, map[string]string{"*": "blah.txt"}},
		{"CONNECT", "/web/blah.txt", true, map[string]string{"*": "blah.txt"}},
		{"OPTIONS", "/web/blah.txt", true, map[string]string{"*": "blah.txt"}},
		{"HEAD", "/web/blah.txt", true, map[string]string{"*": "blah.txt"}},

		{"GET", "/other/1234", true, map[string]string{"id": "1234"}},
	} {
		ctx := chi.NewRouteContext()
		matched := s.r.Match(ctx, c.method, c.path)

		if matched != c.wantMatch {
			t.Errorf("case %d: match: expected %t, got %t", i, c.wantMatch, matched)
			continue
		}

		if c.wantMatch {
			for k, want := range c.wantParams {
				if got := ctx.URLParam(k); got != want {
					t.Errorf("case %d: want %q=%q, got %q=%q", i, k, want, k, got)
				}
			}
		}
	}
}
