package wish

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
)

// RequestUtils is an embeddable type that provides several
// useful functions for use in handlers
type RequestUtils struct{}

// Param attempts to look up a parameter on the request from the following
// sources, in order of priority:
// 1. URL Parameter
// 2. Query
// 3. Post Form (if applicable)
func (u RequestUtils) Param(r *http.Request, key string) string {
	if got := chi.URLParam(r, key); got != "" {
		return got
	}

	return r.FormValue(key)
}

// Respond writes data to w with the given status code and content type.
func (u RequestUtils) Respond(w http.ResponseWriter, contentType string, statusCode int, data []byte) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(statusCode)
	w.Write(data)
}

// RespondJSON writes attempts to marshal data to JSON.
// If marshaling succeeds, a json response is written to w,
// otherwise an Internal Server Error is written.
func (u RequestUtils) RespondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		u.RespondError(w, 500)
		return
	}

	u.Respond(w, "application/json", statusCode, jsonData)
}

// RespondError writes an error response to w.
func (u RequestUtils) RespondError(w http.ResponseWriter, statusCode int) {
	u.Respond(w, "text/plain", statusCode, []byte(http.StatusText(statusCode)))
}

// ServeFile will write the contents of filename to w.
func (u RequestUtils) ServeFile(
	w http.ResponseWriter, r *http.Request, filename string,
) {
	http.ServeFile(w, r, filename)
}
