package genie

import (
	"encoding/json"
	"net/http"
)

type UtilityFuncs struct{}

// Respond writes data to w with the given status code and content type.
func (c UtilityFuncs) Respond(w http.ResponseWriter, contentType string, statusCode int, data []byte) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(statusCode)
	w.Write(data)
}

// Error writes an error response to w.
func (c UtilityFuncs) Error(w http.ResponseWriter, statusCode int) {
	c.Respond(w, "text/plain", statusCode, []byte(http.StatusText(statusCode)))
}

// JSON attempts to marshal v to JSON data. If marshaling succeeds, the result is written
// to w with a content type of "application/json" and the status given by statusCode.
// If marshaling fails, a 500 error with a content type of "text/plain" is written to w
// instead.
func (c UtilityFuncs) JSON(w http.ResponseWriter, statusCode int, v interface{}) {
	data, err := json.Marshal(v)
	if err != nil {
		c.Error(w, http.StatusInternalServerError)
		return
	}
	c.Respond(w, "application/json", statusCode, data)
}
