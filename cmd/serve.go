package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/Br0bertinus/callsheet-api/internal/cache"
	"github.com/Br0bertinus/callsheet-api/internal/client"
	"github.com/Br0bertinus/callsheet-api/internal/service"
	"github.com/Br0bertinus/callsheet-api/server"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var addr string

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP server",
	RunE:  runServe,
}

func init() {
	serveCmd.Flags().StringVarP(&addr, "addr", "a", ":8080", "Address to listen on")
}

func runServe(_ *cobra.Command, _ []string) error {
	logger, err := zap.NewProduction()
	if err != nil {
		return fmt.Errorf("failed to initialise logger: %w", err)
	}
	defer logger.Sync()

	apiKey := os.Getenv("TMDB_API_KEY")
	if apiKey == "" {
		logger.Fatal("TMDB_API_KEY environment variable is required")
	}

	// Initialise core dependencies.
	tmdbClient := client.NewTMDBHTTPClient(apiKey)
	creditCache := cache.NewMemoryCache(30 * time.Minute)
	gameSvc := service.NewGameService(tmdbClient, creditCache)

	corsOrigin := os.Getenv("CORS_ORIGIN")
	if corsOrigin == "" {
		corsOrigin = "*"
	}

	srv := server.New(logger, gameSvc, corsOrigin)
	return srv.Start(addr)
}
