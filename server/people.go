package server

import (
	"net/http"
	"strconv"
	"strings"
)

// SearchPeople handles GET /search/people?q=<query>
func (s *Server) SearchPeople(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		s.writeError(w, http.StatusBadRequest, "query parameter 'q' is required")
		return
	}

	actors, err := s.GameService.SearchPeople(query)
	if err != nil {
		s.writeError(w, http.StatusBadGateway, "failed to search people")
		return
	}

	s.writeJSON(w, http.StatusOK, actors)
}

// GetPerson handles GET /people/{id}
func (s *Server) GetPerson(w http.ResponseWriter, r *http.Request) {
	rawID := r.PathValue("id")
	id, err := strconv.Atoi(rawID)
	if err != nil || id <= 0 {
		s.writeError(w, http.StatusBadRequest, "id must be a positive integer")
		return
	}

	actor, err := s.GameService.GetActor(id)
	if err != nil {
		s.writeError(w, http.StatusBadGateway, "failed to fetch person")
		return
	}

	s.writeJSON(w, http.StatusOK, actor)
}
