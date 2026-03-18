package server

import "github.com/Br0bertinus/callsheet-api/internal/domain"

// GameServicer is the interface the Server depends on for all game logic.
// Defining it here (at the point of consumption) rather than in the service
// package follows standard Go interface practice and keeps the server layer
// decoupled from the concrete implementation.
//
//go:generate mockgen -destination=mocks/mock_game_service.go -package=mocks github.com/Br0bertinus/callsheet-api/server GameServicer
type GameServicer interface {
	NewGame(startActorID, targetActorID int) (domain.NewGameResponse, error)
	DailyChallenge() (domain.NewGameResponse, error)
	ValidateStep(req domain.ValidateStepRequest) (domain.ValidateStepResponse, error)
	SearchPeople(query string) ([]domain.Actor, error)
	SearchMovies(query string) ([]domain.Movie, error)
	GetActor(id int) (domain.Actor, error)
}
