package server

import (
	"net/http"
	"strings"
)

// SearchMovies handles GET /search/movies?q=<query>
func (s *Server) SearchMovies(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		s.writeError(w, http.StatusBadRequest, "query parameter 'q' is required")
		return
	}

	movies, err := s.GameService.SearchMovies(query)
	if err != nil {
		s.writeError(w, http.StatusBadGateway, "failed to search movies")
		return
	}

	s.writeJSON(w, http.StatusOK, movies)
}
