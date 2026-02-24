package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Br0bertinus/callsheet-api/internal/cache"
	"github.com/Br0bertinus/callsheet-api/internal/client"
	"github.com/Br0bertinus/callsheet-api/internal/handlers"
	"github.com/Br0bertinus/callsheet-api/internal/service"
)

func main() {
	apiKey := os.Getenv("TMDB_API_KEY")
	if apiKey == "" {
		log.Fatal("TMDB_API_KEY environment variable is required")
	}

	// Wire up dependencies.
	tmdbClient := client.NewTMDBHTTPClient(apiKey)
	creditCache := cache.NewMemoryCache(30 * time.Minute)
	gameSvc := service.NewGameService(tmdbClient, creditCache)

	// Register routes using Go 1.22 ServeMux pattern matching.
	mux := http.NewServeMux()

	peopleHandler := handlers.NewPeopleHandler(gameSvc)
	mux.HandleFunc("GET /search/people", peopleHandler.SearchPeople)
	mux.HandleFunc("GET /people/{id}", peopleHandler.GetPerson)

	gameHandler := handlers.NewGameHandler(gameSvc)
	mux.HandleFunc("POST /game/validate-step", gameHandler.ValidateStep)

	addr := ":8080"
	log.Printf("callsheet-api listening on %s", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
