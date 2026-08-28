package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/hibiken/asynq"
	"github.com/irosadie/open-state/worker/internal/application/use-cases"
	"github.com/irosadie/open-state/worker/internal/infrastructure/queue"
	"github.com/irosadie/open-state/worker/internal/infrastructure/tracing"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	// Worker tracing (PRD §84). No-op when no OTLP endpoint is set.
	ctx := context.Background()
	traceShutdown := tracing.Setup(ctx, tracing.Config{
		OTLPEndpoint: os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
		ServiceName:  envDefault("OTEL_SERVICE_NAME", "openstate-worker"),
	}, logger)
	defer func() { _ = traceShutdown(context.Background()) }()

	srv, err := queue.NewServer()
	if err != nil {
		logger.Error("worker config error", "error", err.Error())
		return
	}

	// Register job handlers
	mux := asynq.NewServeMux()
	mux.Handle(usecases.TypeRuntimeSummary, usecases.NewRuntimeSummaryHandler())

	logger.Info("starting worker")
	if err := srv.Run(mux); err != nil {
		logger.Error("worker error", "error", err.Error())
	}
}

func envDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
