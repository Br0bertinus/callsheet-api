package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/Br0bertinus/callsheet-api/internal/service"
)

// PeopleHandler holds the dependencies for the /search/people and /people/{id} endpoints.
type PeopleHandler struct {
	svc *service.GameService
}

// NewPeopleHandler creates a PeopleHandler.
func NewPeopleHandler(svc *service.GameService) *PeopleHandler {
	return &PeopleHandler{svc: svc}
}

// SearchPeople handles GET /search/people?q=<query>
func (h *PeopleHandler) SearchPeople(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		writeError(w, http.StatusBadRequest, "query parameter 'q' is required")
		return
	}

	actors, err := h.svc.SearchPeople(query)
	if err != nil {
		writeError(w, http.StatusBadGateway, "failed to search people")
		return
	}

	writeJSON(w, http.StatusOK, actors)
}

// GetPerson handles GET /people/{id}
func (h *PeopleHandler) GetPerson(w http.ResponseWriter, r *http.Request) {
	// Go 1.22 ServeMux supports path value extraction via r.PathValue.
	rawID := r.PathValue("id")
	id, err := strconv.Atoi(rawID)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "id must be a positive integer")
		return
	}

	actor, err := h.svc.GetActor(id)
	if err != nil {
		writeError(w, http.StatusBadGateway, "failed to fetch person")
		return
	}

	writeJSON(w, http.StatusOK, actor)
}
