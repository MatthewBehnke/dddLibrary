// Package config loads runtime configuration from the environment, failing
// fast when required values are missing.
package config

import (
	"errors"
	"os"
)

// Config holds the server's runtime configuration.
type Config struct {
	// DatabaseURL is the Postgres connection string (postgres://...).
	DatabaseURL string
	// HTTPAddr is the listen address for the HTTP server.
	HTTPAddr string
}

// Load reads configuration from the environment. DATABASE_URL is required;
// HTTP_ADDR defaults to ":8080".
func Load() (Config, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}

	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	return Config{DatabaseURL: dbURL, HTTPAddr: addr}, nil
}
