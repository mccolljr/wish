package genie

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"github.com/go-chi/chi"
)

type (
	mountFunc func() http.Handler
)

var (
	tHandlerFunc = reflect.TypeOf(http.HandlerFunc(nil))
	tMountFunc   = reflect.TypeOf(mountFunc(nil))
)

// Bootstrap will build a router based on the methods defined on
// the concrete type returned by p.
func Bootstrap(p Provider, mids ...MiddlewareFunc) (*Server, error) {
	ctx, err := p()
	if err != nil {
		return nil, fmt.Errorf("bootstrap: %v", err)
	}

	if ctx == nil {
		return nil, errors.New("bootstrap: nil context")
	}

	ctxV := reflect.ValueOf(ctx)
	ctxT := ctxV.Type()

	if ctxT.Kind() != reflect.Ptr || ctxT.Elem().Kind() != reflect.Struct {
		return nil, fmt.Errorf("bootstrap: expected pointer to a struct, got %s", ctxT)
	}

	type methodEntry struct {
		name  string
		index int
		v     reflect.Value
	}

	handlers := []methodEntry{}
	mounters := []methodEntry{}

	entrySorter := func(list []methodEntry) func(int, int) bool {
		return func(a, b int) bool {
			av, bv := list[a], list[b]
			al, bl := len(av.name), len(bv.name)
			if al == bl {
				return av.name < bv.name
			}
			return al < bl
		}
	}

	for i := 0; i < ctxT.NumMethod(); i++ {
		name := ctxT.Method(i).Name
		v := ctxV.Method(i)
		switch t := v.Type(); true {
		case t.ConvertibleTo(tHandlerFunc):
			handlers = append(handlers, methodEntry{name: name, v: v, index: i})
		case t.ConvertibleTo(tMountFunc):
			mounters = append(mounters, methodEntry{name: name, v: v, index: i})
		}
	}

	r := chi.NewMux()
	r.Use(mids...)

	sort.Slice(handlers, entrySorter(handlers))
	sort.Slice(mounters, entrySorter(mounters))

	for _, e := range handlers {
		method, pat, ok := parseHandler(e.name)
		if !ok {
			continue
		}

		h := reflectHandler{provider: p, index: e.index}
		if method == "HANDLE" {
			r.Handle(pat, h)
		} else {
			r.Method(method, pat, h)
		}
	}

	for _, e := range mounters {
		pat, ok := parseMount(e.name)
		if !ok {
			continue
		}
		h := e.v.Convert(tMountFunc).Interface().(mountFunc)()
		r.Mount(pat, http.StripPrefix(pat, h))
	}

	return &Server{r: r}, nil
}

var handlerRx = regexp.MustCompile(
	`^(Get|Put|Post|Patch|Delete|Trace|Options|Connect|Head|Handle)` +
		`([a-zA-Z][a-zA-Z0-9]*?)(?:By([a-zA-Z][a-zA-Z0-9]*))?$`)

func parseHandler(name string) (prefix, pattern string, ok bool) {
	match := handlerRx.FindStringSubmatch(name)
	if match == nil {
		return "", "", false
	}
	prefix, spec, param := match[1], match[2], match[3]
	sections := parseSections(spec)
	if param != "" {
		sections = append(sections, "{"+strings.ToLower(param)+"}")
	}
	pattern = "/" + strings.Join(sections, "/")
	return strings.ToUpper(prefix), pattern, true
}

var mountRx = regexp.MustCompile(`^Mount([a-zA-Z][a-zA-Z0-9]*)$`)

func parseMount(name string) (pattern string, ok bool) {
	match := mountRx.FindStringSubmatch(name)
	if match == nil {
		return "", false
	}
	spec := match[1]
	sections := parseSections(spec)
	pattern = "/" + strings.Join(sections, "/")
	return pattern, true
}

func parseSections(spec string) []string {
	sections := []string{}
	if len(spec) >= 4 && spec[:4] == "Root" {
		spec = spec[4:]
	}

	if spec != "" {
		isUpper, lastUpper, nextUpper := true, false, unicode.IsUpper(rune(spec[0]))
		sectionStart := 0
		for i := range spec {
			lastUpper = isUpper
			isUpper = nextUpper
			nextUpper = (i >= len(spec)-1) || unicode.IsUpper(rune(spec[i+1]))
			if i > 0 && isUpper {
				if !lastUpper || !nextUpper {
					sections = append(sections, strings.ToLower(spec[sectionStart:i]))
					sectionStart = i
				}
			}
		}

		sections = append(sections, strings.ToLower(spec[sectionStart:]))
	}

	return sections
}
