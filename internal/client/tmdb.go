package client

import "github.com/Br0bertinus/callsheet-api/internal/domain"

// TMDBClient is the boundary between the service layer and the TMDB API.
// Any concrete implementation (real HTTP, test mock) satisfies this interface.
type TMDBClient interface {
	// SearchPeople returns actors matching the query string.
	SearchPeople(query string) ([]domain.Actor, error)

	// GetActor returns a single actor by TMDB person ID.
	GetActor(id int) (domain.Actor, error)

	// GetMovieCredits returns all movies the actor has appeared in.
	GetMovieCredits(actorID int) ([]domain.Movie, error)
}
