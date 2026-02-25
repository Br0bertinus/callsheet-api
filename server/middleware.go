package server

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// responseRecorder wraps http.ResponseWriter to capture the status code written
// by the handler so we can include it in the log line.
type responseRecorder struct {
	http.ResponseWriter
	status int
}

func (rr *responseRecorder) WriteHeader(code int) {
	rr.status = code
	rr.ResponseWriter.WriteHeader(code)
}

// loggingMiddleware logs the method, path, response status, and elapsed time
// for every incoming request.
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rr := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rr, r)

		s.Logger.Info("request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Int("status", rr.status),
			zap.Duration("duration", time.Since(start)),
		)
	})
}
