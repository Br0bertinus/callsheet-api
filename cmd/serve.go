package cmd

import (
	"fmt"

	"github.com/Br0bertinus/callsheet-api/internal/cache"
	"github.com/Br0bertinus/callsheet-api/internal/client"
	"github.com/Br0bertinus/callsheet-api/internal/config"
	"github.com/Br0bertinus/callsheet-api/internal/service"
	"github.com/Br0bertinus/callsheet-api/server"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP server",
	RunE:  runServe,
}

func runServe(_ *cobra.Command, _ []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	logger, err := zap.NewProduction()
	if err != nil {
		return fmt.Errorf("failed to initialise logger: %w", err)
	}
	defer logger.Sync()

	tmdbClient := client.NewTMDBHTTPClient(cfg.TMDB)
	creditCache := cache.NewMemoryCache(cfg.Cache.TTL)
	gameSvc := service.NewGameService(tmdbClient, creditCache, cfg.Game)

	srv := server.New(logger, gameSvc, cfg.CORS.Origin)
	return srv.Start(cfg.Server.Addr, cfg.Server.ShutdownTimeout)
}
