package main

import (
	"log"
	"net/http"
	"os"

	"github.com/mark3labs/mcp-go/server"

	"github.com/irosadie/open-state/api/internal/domain/capability"
	capinfra "github.com/irosadie/open-state/api/internal/infrastructure/capability"
	mcpapi "github.com/irosadie/open-state/api/internal/interfaces/mcp"
)

func main() {
	port := os.Getenv("MCP_PORT")
	if port == "" {
		port = "8030"
	}

	// Build the capability invoker. The persistence-backed repository is not
	// wired in this slice; we use the mock provider (default sandbox, PRD §2064)
	// so the server boots and declares tools. A nil resolver still lets the
	// server start; invoke_capability returns an authorization error until the
	// capability repository is wired (persistence slices).
	invoker := capability.NewCapabilityInvoker(
		nil, // capability resolver not wired in this slice
		capinfra.MockProviderResolver{},
		nil, // schema validator not wired in this slice
		nil, // rate limiter not wired in this slice
		capability.NewInMemoryIdempotencyStore(),
	)

	deps := mcpapi.Dependencies{
		IntentResolver:   stubIntentResolver{},
		CapabilityInvoker: invoker,
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
