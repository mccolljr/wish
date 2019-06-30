package genie

import (
	"encoding/json"
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

// Respond writes data to w with the given status code and content type.
func (c ContextImpl) Respond(w http.ResponseWriter, contentType string, statusCode int, data []byte) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(statusCode)
	w.Write(data)
}

// Error writes an error response to w.
func (c ContextImpl) Error(w http.ResponseWriter, statusCode int) {
	c.Respond(w, "text/plain", statusCode, []byte(http.StatusText(statusCode)))
}

// JSON attempts to marshal v to JSON data. If marshaling succeeds, the result is written
// to w with a content type of "application/json" and the status given by statusCode.
// If marshaling fails, a 500 error with a content type of "text/plain" is written to w
// instead.
func (c ContextImpl) JSON(w http.ResponseWriter, statusCode int, v interface{}) {
	data, err := json.Marshal(v)
	if err != nil {
		c.Error(w, http.StatusInternalServerError)
		return
	}
	c.Respond(w, "application/json", statusCode, data)
}
