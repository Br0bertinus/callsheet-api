package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config is the top-level application configuration.
type Config struct {
	Server ServerConfig `mapstructure:"server"`
	CORS   CORSConfig   `mapstructure:"cors"`
	TMDB   TMDBConfig   `mapstructure:"tmdb"`
	Cache  CacheConfig  `mapstructure:"cache"`
	Game   GameConfig   `mapstructure:"game"`
}

// ServerConfig controls the HTTP server behaviour.
type ServerConfig struct {
	Addr            string        `mapstructure:"addr"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

// CORSConfig controls cross-origin resource sharing.
type CORSConfig struct {
	Origin string `mapstructure:"origin"`
}

// TMDBConfig holds all TMDB API settings.
// APIKey is never read from config.yaml — it must be provided via the
// TMDB_API_KEY environment variable so it is never committed to source control.
type TMDBConfig struct {
	APIKey    string        `mapstructure:"api_key"`
	BaseURL   string        `mapstructure:"base_url"`
	Timeout   time.Duration `mapstructure:"timeout"`
	Endpoints TMDBEndpoints `mapstructure:"endpoints"`
}

// TMDBEndpoints documents every TMDB API path the application consumes.
// Path parameters are written as {id} for readability; the HTTP client
// substitutes real values at runtime.
type TMDBEndpoints struct {
	SearchPerson    string `mapstructure:"search_person"`
	SearchMovie     string `mapstructure:"search_movie"`
	GetPerson       string `mapstructure:"get_person"`
	GetMovieCredits string `mapstructure:"get_movie_credits"`
}

// CacheConfig controls the in-memory credit cache.
type CacheConfig struct {
	TTL time.Duration `mapstructure:"ttl"`
}

// GameConfig controls daily-challenge rollover behaviour.
type GameConfig struct {
	RolloverTimezone string `mapstructure:"rollover_timezone"`
	RolloverHour     int    `mapstructure:"rollover_hour"`
}

// Load reads config.yaml from the working directory, overlays the
// TMDB_API_KEY environment variable, and returns a validated Config.
func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")

	// TMDB_API_KEY is a secret — bind it from the environment only.
	if err := v.BindEnv("tmdb.api_key", "TMDB_API_KEY"); err != nil {
		return nil, fmt.Errorf("binding TMDB_API_KEY: %w", err)
	}

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("reading config.yaml: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	if cfg.TMDB.APIKey == "" {
		return nil, fmt.Errorf("TMDB_API_KEY environment variable is required")
	}

	return &cfg, nil
}
