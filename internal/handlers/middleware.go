package handlers

import (
	"log"
	"net/http"
	"time"
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

// LoggingMiddleware logs the method, path, response status, and elapsed time
// for every incoming request.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rr := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rr, r)

		log.Printf("%s %s -> %d (%s)", r.Method, r.URL.Path, rr.status, time.Since(start))
	})
}
