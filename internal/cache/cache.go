package cache

import "github.com/Br0bertinus/callsheet-api/internal/domain"

// Cache stores actor movie credits keyed by actor ID.
// This avoids redundant TMDB calls for actors seen more than once.
type Cache interface {
	GetCredits(actorID int) ([]domain.Movie, bool)
	SetCredits(actorID int, movies []domain.Movie)
}
