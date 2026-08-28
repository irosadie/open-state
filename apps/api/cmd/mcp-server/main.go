package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/mark3labs/mcp-go/server"

	appservices "github.com/irosadie/open-state/api/internal/application/services"
	"github.com/irosadie/open-state/api/internal/domain/capability"
	"github.com/irosadie/open-state/api/internal/domain/engine"
	capinfra "github.com/irosadie/open-state/api/internal/infrastructure/capability"
	"github.com/irosadie/open-state/api/internal/infrastructure/config"
	infradb "github.com/irosadie/open-state/api/internal/infrastructure/database"
	engineadapter "github.com/irosadie/open-state/api/internal/infrastructure/engineadapter"
	infralog "github.com/irosadie/open-state/api/internal/infrastructure/logging"
	raginfra "github.com/irosadie/open-state/api/internal/infrastructure/rag"
	mcpapi "github.com/irosadie/open-state/api/internal/interfaces/mcp"
)

func main() {
	port := os.Getenv("MCP_PORT")
	if port == "" {
		port = "8030"
	}

	cfg, err := config.Load()
	if err != nil {
		slog.Error("config error", "error", err.Error())
		return
	}

	logger := infralog.New(infralog.Config{Format: cfg.LogFormat, Level: cfg.LogLevel})
	slog.SetDefault(logger)

	ctx := context.Background()
	pool, err := config.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("database error", "error", err.Error())
		return
	}
	defer pool.Close()

	// Composed persistence adapter (ADR-001): provides the repository interfaces
	// the orchestrator and context compiler depend on.
	adapter := infradb.NewPostgresAdapter(pool)

	// Engine adapter + runtime engine: wires the domain state machine into the
	// MCP propose/current-state path (PRD 170).
	engAdapter := engineadapter.New(pool, adapter.Projects(), adapter.Workflows(), adapter.Instances(), adapter.Events())
	runtimeEngine := engine.NewEngine(engAdapter.Repos())

	// Orchestrator service: lifecycle, propose-event, instances, history,
	// allowed capabilities (PRD 25, 38, 42-43, 52, 142). Engine-backed.
	orchestrator := appservices.NewEngineOrchestratorService(
		adapter.Instances(),
		adapter.Events(),
		adapter.Context(),
		adapter.Capabilities(),
		runtimeEngine,
	)

	// Context compiler: minimal per-turn context with PII redaction (PRD 22, 90).
	// A stub RAG provider is wired until a concrete backend lands (PRD 171).
	contextCompiler := appservices.NewContextCompiler(
		adapter.Context(),
		raginfra.StubRAGProvider{},
		raginfra.NewDefaultRedactor(),
	)

	// Capability invoker (sandbox/mock provider by default, PRD §2064). Wired with
	// a repository-backed resolver (authorization, PRD 59-62) and the JSON schema
	// validator (payload validation, PRD 62).
	capResolver := capability.NewCapabilityResolver(adapter.Capabilities())
	invoker := capability.NewCapabilityInvoker(
		capResolver,
		capinfra.MockProviderResolver{},
		capinfra.JSONSchemaValidator{},
		nil, // rate limiter not wired in this slice
		capability.NewInMemoryIdempotencyStore(),
	)

	// Intent service: resolves conversation intents to real workflows (PRD 38, 171).
	intentSvc := appservices.NewIntentService(adapter.Workflows())

	deps := mcpapi.Dependencies{
		IntentResolver:    intentSvc,
		CapabilityInvoker: invoker,
		Orchestrator:      orchestrator,
		ContextCompiler:   contextCompiler,
	}

	srv := mcpapi.NewServer(deps)
	streamable := server.NewStreamableHTTPServer(srv)

	mux := http.NewServeMux()
	mux.Handle("/mcp", streamable)
	mux.HandleFunc("/health/live", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("/health/ready", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ready"}`))
	})

	addr := ":" + port
	logger.Info("MCP server listening", "addr", addr, "endpoint", "/mcp")
	if err := http.ListenAndServe(addr, mux); err != nil {
		logger.Error("MCP server error", "error", err.Error())
	}
}
