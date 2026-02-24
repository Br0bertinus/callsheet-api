package service

import (
	"fmt"

	"github.com/Br0bertinus/callsheet-api/internal/cache"
	"github.com/Br0bertinus/callsheet-api/internal/client"
	"github.com/Br0bertinus/callsheet-api/internal/domain"
)

// GameService handles all game logic: actor lookup, search, and step validation.
type GameService struct {
	tmdb  client.TMDBClient
	cache cache.Cache
}

// NewGameService creates a GameService with the provided TMDB client and cache.
func NewGameService(tmdb client.TMDBClient, cache cache.Cache) *GameService {
	return &GameService{
		tmdb:  tmdb,
		cache: cache,
	}
}

// SearchPeople searches TMDB for actors matching the query string.
func (s *GameService) SearchPeople(query string) ([]domain.Actor, error) {
	if query == "" {
		return nil, fmt.Errorf("query must not be empty")
	}

	return s.tmdb.SearchPeople(query)
}

// GetActor returns a single actor by TMDB person ID.
func (s *GameService) GetActor(id int) (domain.Actor, error) {
	return s.tmdb.GetActor(id)
}

// ValidateStep checks whether moving from currentActorID to nextActorID is legal.
//
// A step is invalid when:
//   - nextActorID appears in visitedActorIDs (actor already used)
//   - the two actors share no movies
//
// When valid, the response includes the movies connecting the two actors.
func (s *GameService) ValidateStep(req domain.ValidateStepRequest) (domain.ValidateStepResponse, error) {
	if alreadyVisited(req.NextActorID, req.VisitedActorIDs) {
		return domain.ValidateStepResponse{Valid: false, ConnectingMovies: []domain.Movie{}}, nil
	}

	currentCredits, err := s.fetchCredits(req.CurrentActorID)
	if err != nil {
		return domain.ValidateStepResponse{}, fmt.Errorf("could not fetch credits for actor %d: %w", req.CurrentActorID, err)
	}

	nextCredits, err := s.fetchCredits(req.NextActorID)
	if err != nil {
		return domain.ValidateStepResponse{}, fmt.Errorf("could not fetch credits for actor %d: %w", req.NextActorID, err)
	}

	shared := intersectMovies(currentCredits, nextCredits)

	return domain.ValidateStepResponse{
		Valid:            len(shared) > 0,
		ConnectingMovies: shared,
	}, nil
}

// fetchCredits returns cached credits when available, otherwise calls TMDB and caches the result.
func (s *GameService) fetchCredits(actorID int) ([]domain.Movie, error) {
	if movies, ok := s.cache.GetCredits(actorID); ok {
		return movies, nil
	}

	movies, err := s.tmdb.GetMovieCredits(actorID)
	if err != nil {
		return nil, err
	}

	s.cache.SetCredits(actorID, movies)

	return movies, nil
}

// alreadyVisited reports whether actorID appears in the visited slice.
func alreadyVisited(actorID int, visited []int) bool {
	for _, id := range visited {
		if id == actorID {
			return true
		}
	}

	return false
}

// intersectMovies returns movies that appear in both a and b, matched by movie ID.
func intersectMovies(a, b []domain.Movie) []domain.Movie {
	// Index b by movie ID for O(n+m) lookup.
	bIndex := make(map[int]domain.Movie, len(b))
	for _, m := range b {
		bIndex[m.ID] = m
	}

	var shared []domain.Movie
	for _, m := range a {
		if _, ok := bIndex[m.ID]; ok {
			shared = append(shared, m)
		}
	}

	if shared == nil {
		return []domain.Movie{}
	}

	return shared
}
