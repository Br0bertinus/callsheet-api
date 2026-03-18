package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/Br0bertinus/callsheet-api/internal/config"
	"github.com/Br0bertinus/callsheet-api/internal/domain"
)

// tmdbPerson represents a person object in TMDB API responses.
type tmdbPerson struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	ProfilePath string  `json:"profile_path"`
	Popularity  float64 `json:"popularity"`
}

func (p tmdbPerson) toDomain() domain.Actor {
	return domain.Actor{
		ID:          p.ID,
		Name:        p.Name,
		ProfilePath: p.ProfilePath,
		Popularity:  p.Popularity,
	}
}

// tmdbMovie represents a movie object in TMDB API responses.
type tmdbMovie struct {
	ID          int     `json:"id"`
	Title       string  `json:"title"`
	ReleaseDate string  `json:"release_date"`
	PosterPath  string  `json:"poster_path"`
	Popularity  float64 `json:"popularity"`
}

func (m tmdbMovie) toDomain() domain.Movie {
	return domain.Movie{
		ID:         m.ID,
		Title:      m.Title,
		Year:       releaseYear(m.ReleaseDate),
		PosterPath: m.PosterPath,
		Popularity: m.Popularity,
	}
}

// tmdbSearchPersonResponse is the envelope for /search/person.
type tmdbSearchPersonResponse struct {
	Results []tmdbPerson `json:"results"`
}

// tmdbSearchMovieResponse is the envelope for /search/movie.
type tmdbSearchMovieResponse struct {
	Results []tmdbMovie `json:"results"`
}

// tmdbMovieCreditsResponse is the envelope for /person/{id}/movie_credits.
type tmdbMovieCreditsResponse struct {
	Cast []tmdbMovie `json:"cast"`
}

// TMDBHTTPClient makes real HTTP requests to the TMDB API.
type TMDBHTTPClient struct {
	cfg        config.TMDBConfig
	httpClient *http.Client
}

// NewTMDBHTTPClient creates a real TMDB client using the provided config.
func NewTMDBHTTPClient(cfg config.TMDBConfig) *TMDBHTTPClient {
	return &TMDBHTTPClient{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

// SearchPeople calls TMDB /search/person and returns matching actors.
func (c *TMDBHTTPClient) SearchPeople(query string) ([]domain.Actor, error) {
	endpoint := c.cfg.BaseURL + c.cfg.Endpoints.SearchPerson + "?query=" + url.QueryEscape(query)

	var response tmdbSearchPersonResponse

	if err := c.get(endpoint, &response); err != nil {
		return nil, err
	}

	actors := make([]domain.Actor, 0, len(response.Results))
	for _, r := range response.Results {
		actors = append(actors, r.toDomain())
	}

	return actors, nil
}

// GetActor calls TMDB /person/{id} and returns a single actor.
func (c *TMDBHTTPClient) GetActor(id int) (domain.Actor, error) {
	path := strings.ReplaceAll(c.cfg.Endpoints.GetPerson, "{id}", strconv.Itoa(id))
	endpoint := c.cfg.BaseURL + path

	var response tmdbPerson

	if err := c.get(endpoint, &response); err != nil {
		return domain.Actor{}, err
	}

	return response.toDomain(), nil
}

// GetMovieCredits calls TMDB /person/{id}/movie_credits and returns the cast filmography.
func (c *TMDBHTTPClient) GetMovieCredits(actorID int) ([]domain.Movie, error) {
	path := strings.ReplaceAll(c.cfg.Endpoints.GetMovieCredits, "{id}", strconv.Itoa(actorID))
	endpoint := c.cfg.BaseURL + path

	var response tmdbMovieCreditsResponse

	if err := c.get(endpoint, &response); err != nil {
		return nil, err
	}

	movies := make([]domain.Movie, 0, len(response.Cast))
	for _, m := range response.Cast {
		movies = append(movies, m.toDomain())
	}

	return movies, nil
}

// SearchMovies calls TMDB /search/movie and returns matching movies.
func (c *TMDBHTTPClient) SearchMovies(query string) ([]domain.Movie, error) {
	endpoint := c.cfg.BaseURL + c.cfg.Endpoints.SearchMovie + "?query=" + url.QueryEscape(query)

	var response tmdbSearchMovieResponse

	if err := c.get(endpoint, &response); err != nil {
		return nil, err
	}

	movies := make([]domain.Movie, 0, len(response.Results))
	for _, r := range response.Results {
		movies = append(movies, r.toDomain())
	}

	return movies, nil
}

// get performs an HTTP GET with Bearer token auth, decodes the JSON body into dst, and returns any error.
func (c *TMDBHTTPClient) get(endpoint string, dst any) error {
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("tmdb request build failed: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("tmdb request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("%w: tmdb returned 404 for %s", ErrNotFound, endpoint)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("tmdb returned status %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
		return fmt.Errorf("tmdb response decode failed: %w", err)
	}

	return nil
}

// releaseYear parses the year out of a TMDB release_date string ("YYYY-MM-DD").
// Returns 0 if the string is empty or unparseable.
func releaseYear(releaseDate string) int {
	if releaseDate == "" {
		return 0
	}

	yearStr := strings.SplitN(releaseDate, "-", 2)[0]
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		return 0
	}

	return year
}
