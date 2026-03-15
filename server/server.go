package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Br0bertinus/callsheet-api/internal/service"
	"go.uber.org/zap"
)

// Server holds the dependencies shared across HTTP handlers.
type Server struct {
	Logger      *zap.Logger
	GameService *service.GameService
	CORSOrigin  string
}

// New creates a Server with the provided dependencies.
func New(logger *zap.Logger, gameSvc *service.GameService, corsOrigin string) *Server {
	return &Server{
		Logger:      logger,
		GameService: gameSvc,
		CORSOrigin:  corsOrigin,
	}
}

// routes registers all application routes and returns the handler.
func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", s.Health)
	mux.HandleFunc("POST /game", s.NewGame)
	mux.HandleFunc("GET /search/people", s.SearchPeople)
	mux.HandleFunc("GET /search/movies", s.SearchMovies)
	mux.HandleFunc("GET /people/{id}", s.GetPerson)
	mux.HandleFunc("GET /game/daily", s.DailyChallenge)
	mux.HandleFunc("POST /game/validate-step", s.ValidateStep)

	return s.corsMiddleware(s.loggingMiddleware(mux))
}

// Start begins listening for HTTP requests on the given address and blocks until
// an interrupt or SIGTERM signal is received, then shuts down gracefully.
func (s *Server) Start(addr string) error {
	h := &http.Server{
		Addr:    addr,
		Handler: s.routes(),
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sig)

	go func() {
		s.Logger.Info("callsheet-api listening", zap.String("addr", addr))
		if err := h.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.Logger.Fatal("server error", zap.Error(err))
		}
	}()

	<-sig

	s.Logger.Info("shutting down the server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := h.Shutdown(ctx); err != nil {
		return err
	}

	s.Logger.Info("server gracefully shut down")
	return nil
}
