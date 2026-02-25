package server

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

// writeJSON encodes v as JSON and writes it to w with the given status code.
func (s *Server) writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		s.Logger.Error("failed to encode response", zap.Error(err))
	}
}

// writeError writes a plain JSON error message with the given status code.
func (s *Server) writeError(w http.ResponseWriter, status int, message string) {
	s.writeJSON(w, status, map[string]string{"error": message})
}
