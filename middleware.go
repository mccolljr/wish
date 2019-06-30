package genie

import (
	"net/http"

	"github.com/go-chi/chi/middleware"
)

// MiddlewareFunc wraps an http.Handler and provides
// middleware behavior.
type MiddlewareFunc = func(http.Handler) http.Handler

// pre-defined middleware
var (
	DoLog      = middleware.DefaultLogger
	DoRecover  = middleware.Recoverer
	DoCompress = middleware.DefaultCompress
)
