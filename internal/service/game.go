package service

import (
	"fmt"
	"hash/fnv"
	"math/rand"
	"time"

	"github.com/Br0bertinus/callsheet-api/internal/cache"
	"github.com/Br0bertinus/callsheet-api/internal/client"
	"github.com/Br0bertinus/callsheet-api/internal/config"
	"github.com/Br0bertinus/callsheet-api/internal/domain"
)

// GameService handles all game logic: actor lookup, search, and step validation.
type GameService struct {
	tmdb             client.TMDBClient
	cache            cache.Cache
	rolloverTimezone string
	rolloverHour     int
}

// NewGameService creates a GameService with the provided TMDB client, cache, and game config.
func NewGameService(tmdb client.TMDBClient, cache cache.Cache, cfg config.GameConfig) *GameService {
	return &GameService{
		tmdb:             tmdb,
		cache:            cache,
		rolloverTimezone: cfg.RolloverTimezone,
		rolloverHour:     cfg.RolloverHour,
	}
}

// SearchPeople searches TMDB for actors matching the query string.
func (s *GameService) SearchPeople(query string) ([]domain.Actor, error) {
	if query == "" {
		return nil, fmt.Errorf("query must not be empty")
	}

	return s.tmdb.SearchPeople(query)
}

// SearchMovies searches TMDB for movies matching the query string.
func (s *GameService) SearchMovies(query string) ([]domain.Movie, error) {
	if query == "" {
		return nil, fmt.Errorf("query must not be empty")
	}

	return s.tmdb.SearchMovies(query)
}

// GetActor returns a single actor by TMDB person ID.
func (s *GameService) GetActor(id int) (domain.Actor, error) {
	return s.tmdb.GetActor(id)
}

// NewGame validates that both actor IDs exist in TMDB and returns their details
// to bootstrap a game session on the client.
func (s *GameService) NewGame(startActorID, targetActorID int) (domain.NewGameResponse, error) {
	startActor, err := s.tmdb.GetActor(startActorID)
	if err != nil {
		return domain.NewGameResponse{}, fmt.Errorf("could not fetch start actor %d: %w", startActorID, err)
	}

	targetActor, err := s.tmdb.GetActor(targetActorID)
	if err != nil {
		return domain.NewGameResponse{}, fmt.Errorf("could not fetch target actor %d: %w", targetActorID, err)
	}

	return domain.NewGameResponse{
		StartActor:  startActor,
		TargetActor: targetActor,
	}, nil
}

// gameDate returns the current "game day" string (YYYY-MM-DD).
//
// The daily challenge rolls over at the configured hour in the configured
// timezone (default: 01:00 America/Los_Angeles) rather than midnight UTC so
// that US users always see the new puzzle after a reasonable overnight gap.
// Concretely: we shift the clock back by rolloverHour hours before extracting
// the date, so anything before that hour is still treated as the previous
// calendar day.
//
// If the configured timezone cannot be loaded (e.g. tzdata is missing from the
// runtime image) the function falls back to a fixed UTC-8 offset (PST without
// DST adjustment).
func (s *GameService) gameDate() string {
	loc, err := time.LoadLocation(s.rolloverTimezone)
	if err != nil {
		// Fallback: fixed UTC-8 (PST, no DST)
		loc = time.FixedZone("PST", -8*60*60)
	}

	// Subtract rolloverHour hours so the rollover occurs at that hour, not midnight.
	t := time.Now().In(loc).Add(-time.Duration(s.rolloverHour) * time.Hour)
	return t.Format("2006-01-02")
}

// DailyChallenge returns the fixed start/target actor pair for the current game day.
//
// Override priority (highest to lowest):
//  1. dailyOverrides map in daily_overrides.go — planned editorial overrides.
//  2. Seeded PRNG derived from the game-day date string — default behaviour.
func (s *GameService) DailyChallenge() (domain.NewGameResponse, error) {
	dateStr := s.gameDate()

	// 1. Code-level override map (daily_overrides.go).
	if pair, ok := dailyOverrides[dateStr]; ok {
		return s.NewGame(pair[0], pair[1])
	}

	// 2. Default: deterministic PRNG seeded from the UTC date string.
	h := fnv.New64a()
	h.Write([]byte(dateStr))
	seed := int64(h.Sum64())

	rng := rand.New(rand.NewSource(seed)) //nolint:gosec // seeded PRNG for game logic, not security

	n := len(dailyActorPool)
	i := rng.Intn(n)
	j := rng.Intn(n - 1)
	if j >= i {
		j++ // shift to avoid collision without bias
	}

	return s.NewGame(dailyActorPool[i], dailyActorPool[j])
}

// ValidateStep checks whether a user's proposed chain step is legal.
//
// A step is invalid when:
//   - nextActorID appears in visitedActorIDs (actor already used)
//   - movieID appears in visitedMovieIDs (movie already used in the chain)
//   - the two actors do not share the specified movie
//
// When valid, the response includes all shared movies between the two actors
// so the client can display them as hints or confirm the correct answer.
func (s *GameService) ValidateStep(req domain.ValidateStepRequest) (domain.ValidateStepResponse, error) {
	empty := domain.ValidateStepResponse{Valid: false, ConnectingMovies: []domain.Movie{}}

	if alreadyVisited(req.NextActorID, req.VisitedActorIDs) {
		return empty, nil
	}

	if alreadyVisited(req.MovieID, req.VisitedMovieIDs) {
		return empty, nil
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

	// The user must name the specific movie connecting the two actors.
	namedMovieIsShared := false
	for _, m := range shared {
		if m.ID == req.MovieID {
			namedMovieIsShared = true
			break
		}
	}

	return domain.ValidateStepResponse{
		Valid:            namedMovieIsShared,
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
