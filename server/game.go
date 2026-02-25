package server

import (
	"encoding/json"
	"net/http"

	"github.com/Br0bertinus/callsheet-api/internal/domain"
)

// ValidateStep handles POST /game/validate-step.
func (s *Server) ValidateStep(w http.ResponseWriter, r *http.Request) {
	var req domain.ValidateStepRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.CurrentActorID <= 0 || req.NextActorID <= 0 {
		s.writeError(w, http.StatusBadRequest, "currentActorId and nextActorId must be positive integers")
		return
	}

	result, err := s.GameService.ValidateStep(req)
	if err != nil {
		s.writeError(w, http.StatusBadGateway, "failed to validate step")
		return
	}

	s.writeJSON(w, http.StatusOK, result)
}
