package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Br0bertinus/callsheet-api/internal/domain"
	"github.com/Br0bertinus/callsheet-api/internal/service"
)

// GameHandler holds the dependencies for the /game/* endpoints.
type GameHandler struct {
	svc *service.GameService
}

// NewGameHandler creates a GameHandler.
func NewGameHandler(svc *service.GameService) *GameHandler {
	return &GameHandler{svc: svc}
}

// ValidateStep handles POST /game/validate-step.
// It decodes the request body, delegates to the service, and returns the result.
func (h *GameHandler) ValidateStep(w http.ResponseWriter, r *http.Request) {
	var req domain.ValidateStepRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.CurrentActorID <= 0 || req.NextActorID <= 0 {
		writeError(w, http.StatusBadRequest, "currentActorId and nextActorId must be positive integers")
		return
	}

	result, err := h.svc.ValidateStep(req)
	if err != nil {
		writeError(w, http.StatusBadGateway, "failed to validate step")
		return
	}

	writeJSON(w, http.StatusOK, result)
}
