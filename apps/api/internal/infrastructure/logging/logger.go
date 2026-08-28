package logging

import (
	"log/slog"
	"os"
)

// Config controls the structured logger output.
type Config struct {
	// Format is "json" (default) or "text".
	Format string
	// Level is a slog level string: debug, info, warn, error.
	Level string
}

// New builds a structured logger (log/slog) from config. JSON is the default
// handler for production ingestion; text is available for local development.
func New(cfg Config) *slog.Logger {
	level := parseLevel(cfg.Level)

	var handler slog.Handler
	if cfg.Format == "text" {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	} else {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	}

	return slog.New(handler)
}

func parseLevel(s string) slog.Level {
	switch s {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
