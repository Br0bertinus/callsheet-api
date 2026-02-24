package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Br0bertinus/callsheet-api/internal/domain"
)

const tmdbBaseURL = "https://api.themoviedb.org/3"

// TMDBHTTPClient makes real HTTP requests to the TMDB API.
type TMDBHTTPClient struct {
	apiKey     string
	httpClient *http.Client
}

// NewTMDBHTTPClient creates a real TMDB client using the provided API key.
func NewTMDBHTTPClient(apiKey string) *TMDBHTTPClient {
	return &TMDBHTTPClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SearchPeople calls TMDB /search/person and returns matching actors.
func (c *TMDBHTTPClient) SearchPeople(query string) ([]domain.Actor, error) {
	endpoint := fmt.Sprintf("%s/search/person?query=%s&api_key=%s",
		tmdbBaseURL,
		url.QueryEscape(query),
		c.apiKey,
	)

	var response struct {
		Results []struct {
			ID          int    `json:"id"`
			Name        string `json:"name"`
			ProfilePath string `json:"profile_path"`
		} `json:"results"`
	}

	if err := c.get(endpoint, &response); err != nil {
		return nil, err
	}

	actors := make([]domain.Actor, 0, len(response.Results))
	for _, r := range response.Results {
		actors = append(actors, domain.Actor{
			ID:          r.ID,
			Name:        r.Name,
			ProfilePath: r.ProfilePath,
		})
	}

	return actors, nil
}

// GetActor calls TMDB /person/{id} and returns a single actor.
func (c *TMDBHTTPClient) GetActor(id int) (domain.Actor, error) {
	endpoint := fmt.Sprintf("%s/person/%d?api_key=%s", tmdbBaseURL, id, c.apiKey)

	var response struct {
		ID          int    `json:"id"`
		Name        string `json:"name"`
		ProfilePath string `json:"profile_path"`
	}

	if err := c.get(endpoint, &response); err != nil {
		return domain.Actor{}, err
	}

	return domain.Actor{
		ID:          response.ID,
		Name:        response.Name,
		ProfilePath: response.ProfilePath,
	}, nil
}

// GetMovieCredits calls TMDB /person/{id}/movie_credits and returns the cast filmography.
func (c *TMDBHTTPClient) GetMovieCredits(actorID int) ([]domain.Movie, error) {
	endpoint := fmt.Sprintf("%s/person/%d/movie_credits?api_key=%s", tmdbBaseURL, actorID, c.apiKey)

	var response struct {
		Cast []struct {
			ID          int    `json:"id"`
			Title       string `json:"title"`
			ReleaseDate string `json:"release_date"`
		} `json:"cast"`
	}

	if err := c.get(endpoint, &response); err != nil {
		return nil, err
	}

	movies := make([]domain.Movie, 0, len(response.Cast))
	for _, m := range response.Cast {
		movies = append(movies, domain.Movie{
			ID:    m.ID,
			Title: m.Title,
			Year:  releaseYear(m.ReleaseDate),
		})
	}

	return movies, nil
}

// get performs an HTTP GET, decodes the JSON body into dst, and returns any error.
func (c *TMDBHTTPClient) get(url string, dst any) error {
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("tmdb request failed: %w", err)
	}
	defer resp.Body.Close()

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
