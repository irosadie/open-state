package config

import (
	"fmt"
	"os"
)

// Config holds all environment-sourced configuration.
type Config struct {
	DatabaseURL    string
	Port           string
	JWTSecret      string
	LogFormat      string
	LogLevel       string
	MetricsEnabled bool
	OTel           OTelConfig
	RateLimit      RateLimitConfig
	SSO            SSOConfig
	Security       SecurityConfig
}

// OTelConfig controls OpenTelemetry tracing (PRD §84). When the OTLP endpoint is
// empty, tracing is a no-op and the service starts normally.
type OTelConfig struct {
	OTLPEndpoint string
	ServiceName  string
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
	// Secrets management (PRD §139): enforce a minimum key length so weak secrets
	// are rejected at startup, not silently used.
	if len(jwtSecret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8020"
	}

	return &Config{
		DatabaseURL:    dbURL,
		Port:           port,
		JWTSecret:      jwtSecret,
		LogFormat:      envDefault("LOG_FORMAT", "json"),
		LogLevel:       envDefault("LOG_LEVEL", "info"),
		MetricsEnabled: envBool("METRICS_ENABLED", true),
		OTel: OTelConfig{
			OTLPEndpoint: os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
			ServiceName:  envDefault("OTEL_SERVICE_NAME", "openstate-api"),
		},
		RateLimit: loadRateLimits(),
		SSO:       loadSSO(),
		Security:  loadSecurity(),
	}, nil
}

func envBool(key string, def bool) bool {
	raw := os.Getenv(key)
	if raw == "" {
		return def
	}
	return raw == "1" || raw == "true" || raw == "yes"
}

func envDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
