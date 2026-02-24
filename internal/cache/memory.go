package cache

import (
	"sync"
	"time"

	"github.com/Br0bertinus/callsheet-api/internal/domain"
)

// entry wraps a cached value with an expiry timestamp.
type entry struct {
	movies    []domain.Movie
	expiresAt time.Time
}

// MemoryCache is a simple TTL-based in-memory cache.
// It is safe for concurrent use.
type MemoryCache struct {
	mu    sync.RWMutex
	items map[int]entry
	ttl   time.Duration
}

// NewMemoryCache creates a MemoryCache where each entry expires after ttl.
func NewMemoryCache(ttl time.Duration) *MemoryCache {
	return &MemoryCache{
		items: make(map[int]entry),
		ttl:   ttl,
	}
}

// GetCredits returns the cached movies for actorID, if present and not expired.
func (c *MemoryCache) GetCredits(actorID int) ([]domain.Movie, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	e, ok := c.items[actorID]
	if !ok || time.Now().After(e.expiresAt) {
		return nil, false
	}

	return e.movies, true
}

// SetCredits stores movies for actorID with the configured TTL.
func (c *MemoryCache) SetCredits(actorID int, movies []domain.Movie) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[actorID] = entry{
		movies:    movies,
		expiresAt: time.Now().Add(c.ttl),
	}
}
