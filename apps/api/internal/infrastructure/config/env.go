package config

import (
	"fmt"
	"os"
)

// Config holds all environment-sourced configuration.
type Config struct {
	DatabaseURL string
	Port        string
	JWTSecret   string
	RateLimit   RateLimitConfig
}

// Load reads required env vars and fails fast if any are missing.
func Load() (*Config, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8020"
	}

	return &Config{
		DatabaseURL: dbURL,
		Port:        port,
		JWTSecret:   jwtSecret,
		RateLimit:   loadRateLimits(),
	}, nil
}
