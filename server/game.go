package server

import (
	"encoding/json"
	"net/http"

	"github.com/Br0bertinus/callsheet-api/internal/domain"
)

// NewGame handles POST /game.
// The client supplies the two actor IDs they want to connect; the server confirms
// both exist in TMDB and returns their full details to bootstrap the game session.
func (s *Server) NewGame(w http.ResponseWriter, r *http.Request) {
	var req domain.NewGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.StartActorID <= 0 || req.TargetActorID <= 0 {
		s.writeError(w, http.StatusBadRequest, "startActorId and targetActorId must be positive integers")
		return
	}

	if req.StartActorID == req.TargetActorID {
		s.writeError(w, http.StatusBadRequest, "startActorId and targetActorId must be different")
		return
	}

	result, err := s.GameService.NewGame(req.StartActorID, req.TargetActorID)
	if err != nil {
		s.writeError(w, http.StatusBadGateway, "failed to start game")
		return
	}

	s.writeJSON(w, http.StatusCreated, result)
}

// DailyChallenge handles GET /game/daily.
// It returns the same start/target actor pair for every caller on the same
// UTC calendar day. No authentication required.
func (s *Server) DailyChallenge(w http.ResponseWriter, r *http.Request) {
	result, err := s.GameService.DailyChallenge()
	if err != nil {
		s.writeError(w, http.StatusBadGateway, "failed to fetch daily challenge")
		return
	}

	s.writeJSON(w, http.StatusOK, result)
}

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

	if req.MovieID <= 0 {
		s.writeError(w, http.StatusBadRequest, "movieId must be a positive integer")
		return
	}

	result, err := s.GameService.ValidateStep(req)
	if err != nil {
		s.writeError(w, http.StatusBadGateway, "failed to validate step")
		return
	}

	s.writeJSON(w, http.StatusOK, result)
}
