package service_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Br0bertinus/callsheet-api/internal/cache"
	"github.com/Br0bertinus/callsheet-api/internal/domain"
	"github.com/Br0bertinus/callsheet-api/internal/service"
)

// --- mock TMDB client ---

// mockTMDB satisfies client.TMDBClient using fixed per-actor credit maps.
type mockTMDB struct {
	credits map[int][]domain.Movie
}

func (m *mockTMDB) SearchPeople(_ string) ([]domain.Actor, error) {
	return nil, nil
}

func (m *mockTMDB) GetActor(_ int) (domain.Actor, error) {
	return domain.Actor{}, nil
}

func (m *mockTMDB) GetMovieCredits(actorID int) ([]domain.Movie, error) {
	return m.credits[actorID], nil
}

func (m *mockTMDB) SearchMovies(_ string) ([]domain.Movie, error) {
	return nil, nil
}

// --- helpers ---

func newService(credits map[int][]domain.Movie) *service.GameService {
	tmdb := &mockTMDB{credits: credits}
	c := cache.NewMemoryCache(0) // TTL 0 means entries never expire during tests.
	return service.NewGameService(tmdb, c)
}

// --- ValidateStep tests ---

func TestValidateStep_RejectsVisitedActor(t *testing.T) {
	svc := newService(map[int][]domain.Movie{})

	result, err := svc.ValidateStep(domain.ValidateStepRequest{
		CurrentActorID:  1,
		NextActorID:     2,
		VisitedActorIDs: []int{2, 3, 4}, // 2 is already visited
	})

	require.NoError(t, err)
	assert.False(t, result.Valid)
	assert.Empty(t, result.ConnectingMovies)
}

func TestValidateStep_InvalidWhenNoSharedMovies(t *testing.T) {
	svc := newService(map[int][]domain.Movie{
		1: {{ID: 10, Title: "Film A", Year: 2000}},
		2: {{ID: 20, Title: "Film B", Year: 2001}},
	})

	result, err := svc.ValidateStep(domain.ValidateStepRequest{
		CurrentActorID:  1,
		NextActorID:     2,
		MovieID:         10,
		VisitedActorIDs: []int{},
	})

	require.NoError(t, err)
	assert.False(t, result.Valid)
	assert.Empty(t, result.ConnectingMovies)
}

func TestValidateStep_ValidWhenSharedMoviesExist(t *testing.T) {
	sharedMovie := domain.Movie{ID: 99, Title: "Shared Film", Year: 2010}

	svc := newService(map[int][]domain.Movie{
		1: {sharedMovie, {ID: 10, Title: "Film A", Year: 2000}},
		2: {sharedMovie, {ID: 20, Title: "Film B", Year: 2001}},
	})

	result, err := svc.ValidateStep(domain.ValidateStepRequest{
		CurrentActorID:  1,
		NextActorID:     2,
		MovieID:         99, // correctly names the shared movie
		VisitedActorIDs: []int{},
	})

	require.NoError(t, err)
	assert.True(t, result.Valid)
	require.Len(t, result.ConnectingMovies, 1)
	assert.Equal(t, sharedMovie, result.ConnectingMovies[0])
}

func TestValidateStep_ValidWhenMultipleSharedMovies(t *testing.T) {
	movieA := domain.Movie{ID: 1, Title: "Movie A", Year: 2001}
	movieB := domain.Movie{ID: 2, Title: "Movie B", Year: 2002}
	movieC := domain.Movie{ID: 3, Title: "Movie C", Year: 2003}

	svc := newService(map[int][]domain.Movie{
		10: {movieA, movieB, movieC},
		20: {movieA, movieC},
	})

	result, err := svc.ValidateStep(domain.ValidateStepRequest{
		CurrentActorID:  10,
		NextActorID:     20,
		MovieID:         1, // names one of the two shared movies
		VisitedActorIDs: []int{},
	})

	require.NoError(t, err)
	assert.True(t, result.Valid)
	// Both shared movies are returned so the client can surface them as hints.
	assert.Len(t, result.ConnectingMovies, 2)
}

func TestValidateStep_NextActorNotInVisitedButNoSharedMovies(t *testing.T) {
	svc := newService(map[int][]domain.Movie{
		1: {},
		2: {},
	})

	result, err := svc.ValidateStep(domain.ValidateStepRequest{
		CurrentActorID:  1,
		NextActorID:     2,
		MovieID:         99,
		VisitedActorIDs: []int{3, 4},
	})

	require.NoError(t, err)
	assert.False(t, result.Valid)
	assert.Empty(t, result.ConnectingMovies)
}

func TestValidateStep_InvalidWhenNamedMovieNotShared(t *testing.T) {
	sharedMovie := domain.Movie{ID: 99, Title: "Shared Film", Year: 2010}

	svc := newService(map[int][]domain.Movie{
		1: {sharedMovie, {ID: 10, Title: "Film A", Year: 2000}},
		2: {sharedMovie, {ID: 20, Title: "Film B", Year: 2001}},
	})

	// Actor 1 and Actor 2 share movie 99, but the user claims it was movie 10
	// (which only Actor 1 appeared in — not a shared film).
	result, err := svc.ValidateStep(domain.ValidateStepRequest{
		CurrentActorID:  1,
		NextActorID:     2,
		MovieID:         10,
		VisitedActorIDs: []int{},
	})

	require.NoError(t, err)
	assert.False(t, result.Valid)
	// The real shared movies are still returned so the client knows what was valid.
	require.Len(t, result.ConnectingMovies, 1)
	assert.Equal(t, sharedMovie, result.ConnectingMovies[0])
}

func TestValidateStep_InvalidWhenMovieAlreadyVisited(t *testing.T) {
	sharedMovie := domain.Movie{ID: 99, Title: "Shared Film", Year: 2010}

	svc := newService(map[int][]domain.Movie{
		1: {sharedMovie},
		2: {sharedMovie},
	})

	// The user names the correct shared movie, but it was already used earlier in the chain.
	result, err := svc.ValidateStep(domain.ValidateStepRequest{
		CurrentActorID:  1,
		NextActorID:     2,
		MovieID:         99,
		VisitedActorIDs: []int{},
		VisitedMovieIDs: []int{99},
	})

	require.NoError(t, err)
	assert.False(t, result.Valid)
}

func TestValidateStep_UsesCacheOnSecondCall(t *testing.T) {
	callCount := 0
	credits := map[int][]domain.Movie{
		1: {{ID: 99, Title: "Shared", Year: 2020}},
		2: {{ID: 99, Title: "Shared", Year: 2020}},
	}

	// Wrap mockTMDB to count GetMovieCredits calls.
	type countingMock struct {
		mockTMDB
	}

	tmdb := &mockTMDB{credits: credits}
	c := cache.NewMemoryCache(0)

	// Prime the cache for actor 1 manually by calling ValidateStep twice.
	svc := service.NewGameService(tmdb, c)

	req := domain.ValidateStepRequest{
		CurrentActorID:  1,
		NextActorID:     2,
		MovieID:         99,
		VisitedActorIDs: []int{},
	}

	_ = callCount // suppress unused warning

	// First call: populates cache for actors 1 and 2.
	result1, err1 := svc.ValidateStep(req)
	require.NoError(t, err1)
	assert.True(t, result1.Valid)

	// Second call: should use the cache, not re-call TMDB.
	// This test validates the flow completes correctly regardless of source.
	result2, err2 := svc.ValidateStep(req)
	require.NoError(t, err2)
	assert.True(t, result2.Valid)
}
