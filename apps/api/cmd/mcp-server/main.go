package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/mark3labs/mcp-go/server"

	appservices "github.com/irosadie/open-state/api/internal/application/services"
	"github.com/irosadie/open-state/api/internal/domain/capability"
	domaincap "github.com/irosadie/open-state/api/internal/domain/capability"
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
		adapter.Workflows(),
		adapter.CapabilityEvidence(),
	)
	orchestrator.SetProjectCapabilityMCPBindings(adapter.ProjectMCPBindings())

	// Context compiler: minimal per-turn context with PII redaction (PRD 22, 90).
	// A stub RAG provider is wired until a concrete backend lands (PRD 171).
	contextCompiler := appservices.NewContextCompiler(
		adapter.Context(),
		raginfra.StubRAGProvider{},
		raginfra.NewDefaultRedactor(),
	)
	traceRecorder := appservices.NewRuntimeTraceRecorder(adapter.RuntimeTraces())
	apiKeySvc := appservices.NewAPIKeyService(adapter.APIKeys(), adapter.Projects(), nil, cfg.MCPAPIKeyPepper)

	// Capability invoker (sandbox/mock provider by default, PRD §2064). Wired with
	// a repository-backed resolver (authorization, PRD 59-62) and the JSON schema
	// validator (payload validation, PRD 62).
	// When FIXTURE_FILE is set, use the JSON-file provider for flow testing.
	capResolver := capability.NewCapabilityResolver(adapter.Capabilities())
	var providerResolver domaincap.ProviderResolver
	if fixturePath := os.Getenv("FIXTURE_FILE"); fixturePath != "" {
		fr, err := capinfra.NewJSONFileProviderResolver(fixturePath)
		if err != nil {
			logger.Error("fixture provider error", "error", err.Error(), "path", fixturePath)
			return
		}
		logger.Info("using JSON fixture provider", "path", fixturePath)
		providerResolver = fr
	} else {
		providerResolver = capinfra.MockProviderResolver{}
	}
	invoker := capability.NewCapabilityInvoker(
		capResolver,
		providerResolver,
		capinfra.JSONSchemaValidator{},
		nil, // rate limiter not wired in this slice
		capability.NewInMemoryIdempotencyStore(),
	)

	// The secure gateway uses the same trusted transport adapter as the MCP
	// connection admin service. It resolves the connection and exact discovered
	// tool from project bindings before this adapter can call a provider.
	mcpGatewayProvider := capinfra.NewMCPConnectionTester(nil, capinfra.EnvCredentialResolver{Prefix: "CRED_"}, 10*time.Second)
	mcpGateway := appservices.NewMCPGatewayService(
		orchestrator,
		adapter.Capabilities(),
		adapter.ProjectMCPBindings(),
		adapter.MCPConnections(),
		adapter.MCPToolCatalog(),
		adapter.CapabilityEvidence(),
		adapter.Context(),
		adapter.Workflows(),
		mcpGatewayProvider,
		capinfra.JSONSchemaValidator{},
		invoker,
		10*time.Second,
	)

	// Intent service: resolves conversation intents to real workflows (PRD 38, 171).
	intentSvc := appservices.NewIntentService(adapter.Intents(), adapter.Workflows())

	deps := mcpapi.Dependencies{
		APIKeyAuth:                apiKeySvc,
		IntentResolver:            intentSvc,
		CapabilityInvoker:         invoker,
		Orchestrator:              orchestrator,
		ContextCompiler:           contextCompiler,
		TraceRecorder:             traceRecorder,
		ContextRepo:               adapter.Context(),
		CapabilityRegistry:        adapter.Capabilities(),
		CapabilityEvidence:        adapter.CapabilityEvidence(),
		ProjectCapabilityBindings: adapter.ProjectMCPBindings(),
		WorkflowRegistry:          adapter.Workflows(),
		CapabilityOutputValidator: capinfra.JSONSchemaValidator{},
		Gateway:                   mcpGateway,
		GatewayMode:               appservices.MCPGatewayMode(cfg.MCPGatewayMode),
	}

	srv := mcpapi.NewServer(deps)
	streamable := server.NewStreamableHTTPServer(srv)

	mux := http.NewServeMux()
	mux.Handle("/mcp", mcpapi.APIKeyAuthentication(apiKeySvc)(streamable))
	mux.HandleFunc("/health/live", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("/health/ready", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ready","server":"openstate"}`))
	})

	addr := ":" + port
	logger.Info("MCP server listening", "addr", addr, "endpoint", "/mcp")
	if err := http.ListenAndServe(addr, mux); err != nil {
		logger.Error("MCP server error", "error", err.Error())
	}
}
