package genie

import (
	"net/http"

	"github.com/go-chi/chi"
)

// A Provider is a function that will return a Context
// to be used for handling a server request.
// A Provider must always return values of the same
// concrete type.
type Provider func() (Context, error)

// Context represents a type that can be used to handle a
// server request.
type Context interface {
	context()
}

// ContextImpl must be embedded in all types that implement Context.
// It provides several utility methods for use in handlers.
type ContextImpl struct{}

func (*ContextImpl) context() {}

// Param attempts to look up a parameter on the request from the following
// sources, in order of priority:
// 1. URL Parameter
// 2. Query
// 3. Post Form (if applicable)
func (c ContextImpl) Param(r *http.Request, key string) string {
	if got := chi.URLParam(r, key); got != "" {
		return got
	}

	return r.FormValue(key)
}
