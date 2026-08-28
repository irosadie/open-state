package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/mark3labs/mcp-go/server"

	appservices "github.com/irosadie/open-state/api/internal/application/services"
	"github.com/irosadie/open-state/api/internal/domain/capability"
	capinfra "github.com/irosadie/open-state/api/internal/infrastructure/capability"
	"github.com/irosadie/open-state/api/internal/infrastructure/config"
	infradb "github.com/irosadie/open-state/api/internal/infrastructure/database"
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
		log.Fatalf("config error: %v", err)
	}

	ctx := context.Background()
	pool, err := config.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database error: %v", err)
	}
	defer pool.Close()

	// Composed persistence adapter (ADR-001): provides the repository interfaces
	// the orchestrator and context compiler depend on.
	adapter := infradb.NewPostgresAdapter(pool)

	// Orchestrator service: lifecycle, propose-event, instances, history,
	// allowed capabilities (PRD 25, 38, 42-43, 52, 142).
	orchestrator := appservices.NewOrchestratorService(
		adapter.Instances(),
		adapter.Events(),
		adapter.Context(),
		adapter.Capabilities(),
	)

	// Context compiler: minimal per-turn context with PII redaction (PRD 22, 90).
	// A stub RAG provider is wired until a concrete backend lands (PRD 171).
	contextCompiler := appservices.NewContextCompiler(
		adapter.Context(),
		raginfra.StubRAGProvider{},
		raginfra.NewDefaultRedactor(),
	)

	// Capability invoker (sandbox/mock provider by default, PRD §2064).
	invoker := capability.NewCapabilityInvoker(
		nil, // capability resolver not wired in this slice
		capinfra.MockProviderResolver{},
		nil, // schema validator not wired in this slice
		nil, // rate limiter not wired in this slice
		capability.NewInMemoryIdempotencyStore(),
	)

	deps := mcpapi.Dependencies{
		IntentResolver:  stubIntentResolver{},
		CapabilityInvoker: invoker,
		Orchestrator:     orchestrator,
		ContextCompiler:  contextCompiler,
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
	log.Printf("MCP server listening on %s (Streamable HTTP at /mcp)", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("MCP server error: %v", err)
	}
}

// stubIntentResolver provides intent listing when the real registry is not
// wired in this slice.
type stubIntentResolver struct{}

func (stubIntentResolver) ListIntents() []mcpapi.IntentInfo { return nil }
