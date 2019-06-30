package wish

import (
	"net/http"

	"github.com/go-chi/chi"
)

// A Server is returned by Bootstrap to handle HTTP requests using
// the Context type returned by the Provider passed to Bootstrap.
type Server struct {
	r *chi.Mux
}

// Routes returns a list of all patterns that have
// a registered handler for at least one HTTP method.
func (s *Server) Routes() []string {
	routes := s.r.Routes()
	result := make([]string, len(routes))
	for i, r := range routes {
		result[i] = r.Pattern
	}
	return result
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.r == nil {
		http.Error(w, http.StatusText(404), 404)
		return
	}
	s.r.ServeHTTP(w, r)
}
