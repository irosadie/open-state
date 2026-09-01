// Package mcp implements the MCP server runtime for OpenState (Epic #4,
// mcp-server-runtime). It exposes intent resolution, active workflow lookup,
// compiled context, and authorized capability invocation to external LLM/RAG
// systems over Streamable HTTP (PRD §20, §177).
//
// The core engine stays MCP-agnostic (PRD §172); this package is the interface
// layer that adapts the domain services to the MCP SDK.
package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/irosadie/open-state/api/internal/application/dtos"
	appservices "github.com/irosadie/open-state/api/internal/application/services"
	"github.com/irosadie/open-state/api/internal/domain/capability"
	"github.com/irosadie/open-state/api/internal/domain/engine"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
)

// Dependencies bundles the domain services the MCP server needs. This keeps
// the package decoupled from concrete infrastructure wiring.
type Dependencies struct {
	// APIKeyAuth authenticates State MCP machine principals and writes safe
	// authorization audit entries. It is required for every production tool call.
	APIKeyAuth *appservices.APIKeyService
	// IntentResolver resolves an intent to its workflow (PRD 38, 171).
	IntentResolver IntentPort
	// CapabilityInvoker executes authorized capabilities.
	CapabilityInvoker *capability.CapabilityInvoker
	// Orchestrator exposes runtime workflow orchestration (lifecycle, history,
	// event proposal, instances, allowed capabilities).
	Orchestrator OrchestratorPort
	// ContextCompiler compiles minimal per-turn context (PRD 22).
	ContextCompiler ContextCompilerPort
	// TraceRecorder records application-observed runtime boundaries. It is
	// optional so integrations can remain read/command-only during rollout.
	TraceRecorder *appservices.RuntimeTraceRecorder
	// ContextRepo persists capability response data into workflow instance
	// context so downstream capabilities can read it (PRD §24, §31).
	ContextRepo repositories.IContextRepository
	// CapabilityRegistry resolves logical state capabilities to safe provider
	// server aliases and concrete tools.
	CapabilityRegistry repositories.ICapabilityRepository
	// CapabilityEvidence persists State MCP's accepted provider execution reports.
	CapabilityEvidence repositories.ICapabilityEvidenceRepository
	// ProjectCapabilityBindings resolves logical capabilities to the active,
	// project-scoped discovered MCP tool. Legacy capability provider fields are
	// not a runtime fallback when this mapping is absent.
	ProjectCapabilityBindings repositories.IProjectCapabilityMCPBindingRepository
	// WorkflowRegistry resolves the project for a persisted workflow instance
	// when the orchestration adapter returns a legacy instance projection.
	WorkflowRegistry repositories.IWorkflowRepository
	// CapabilityOutputValidator validates normalized provider results when a
	// capability declares an output schema.
	CapabilityOutputValidator capability.InputSchemaValidator
	// Gateway is the server-side provider execution path used in secure mode.
	Gateway *appservices.MCPGatewayService
	// GatewayMode controls the externally advertised execution contract. Empty
	// keeps the compatibility-safe advisory mode.
	GatewayMode appservices.MCPGatewayMode
}

// IntentPort resolves conversation intents to workflows.
type IntentPort interface {
	ListIntents(ctx context.Context, tenantID, projectID string) ([]entities.Intent, error)
	ResolveIntent(ctx context.Context, tenantID, projectID, intentID string) (*entities.Workflow, error)
}

// OrchestratorPort is the application-layer seam the orchestrator tools delegate
// to. It matches OrchestratorService's public methods.
type OrchestratorPort interface {
	StartWorkflow(ctx context.Context, tenantID, workflowID, workflowVersionID, correlationKey string) (*entities.WorkflowInstance, error)
	SuspendWorkflow(ctx context.Context, tenantID, instanceID string) (*entities.WorkflowInstance, error)
	ResumeWorkflow(ctx context.Context, tenantID, instanceID string) (*entities.WorkflowInstance, error)
	CancelWorkflow(ctx context.Context, tenantID, instanceID string) (*entities.WorkflowInstance, error)
	ListInstances(ctx context.Context, tenantID string) ([]entities.WorkflowInstance, error)
	GetCurrentState(ctx context.Context, tenantID, instanceID string) (*entities.WorkflowInstance, *entities.StateInstance, error)
	GetActiveWorkflow(ctx context.Context, tenantID, conversationID string) (*entities.WorkflowInstance, error)
	GetAllowedTransitions(ctx context.Context, tenantID, instanceID string) ([]engine.TransitionDefinition, error)
	ListHistory(ctx context.Context, tenantID, instanceID string) ([]entities.Event, error)
	ReplayWorkflow(ctx context.Context, tenantID, instanceID string) (map[string]any, *entities.Event, error)
	ProposeEvent(ctx context.Context, tenantID, instanceID, eventType string, payload map[string]any, correlationID string) (*entities.Event, error)
	CurrentStateInfo(ctx context.Context, tenantID, instanceID string) (*engine.StateInfo, error)
	ReplayState(ctx context.Context, tenantID, instanceID string) (map[string]any, string, error)
	ListAllowedCapabilities(ctx context.Context, tenantID string, scopeType entities.BindingScopeType, scopeID string) ([]entities.Capability, error)
}

// ContextCompilerPort is the application-layer seam the get_context tool delegates to.
type ContextCompilerPort interface {
	Compile(ctx context.Context, args appservices.CompileArgs) (*dtos.CompiledContext, error)
}

// IntentInfo is a lightweight projection of an intent for the MCP tool.
type IntentInfo struct {
	ID           string   `json:"id"`
	ProjectID    string   `json:"projectId"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Examples     []string `json:"examples"`
	WorkflowSlug string   `json:"workflowSlug"`
}

// NewServer builds an MCP server with the standard OpenState toolset.
func NewServer(deps Dependencies) *server.MCPServer {
	mode := deps.GatewayMode
	if mode == "" {
		mode = appservices.MCPGatewayModeAdvisory
	}
	instructions := `OpenState is the mandatory state controller and gatekeeper for every configured intent and workflow. Always call list_intents and resolve_intent before starting work, then call get_current_state before proposing a transition. If requiredCapabilities contains a provider requirement, use only the listed project-scoped provider server alias and exact tool name, report the normalized result to report_capability_result, and wait for the gatekeeper response before proposing a transition. A MISSING_MAPPING or UNAVAILABLE requirement is a hard stop: do not guess an endpoint, server, tool, or substitute another MCP. Provider endpoints and credentials are never returned by OpenState and remain managed by the MCP host.`
	if mode == appservices.MCPGatewayModeSecure {
		instructions = `OpenState is the mandatory state controller and enforced gateway for every configured intent and workflow. Always call list_intents and resolve_intent before starting work, then call get_current_state. When the current state declares a capability, call invoke_capability with only the workflow instance, logical capability, correlationId, idempotencyKey, and capability input. OpenState resolves the project binding and calls the provider internally. Never connect to a provider MCP directly, never select a provider server or tool, and never guess a replacement. A MISSING_MAPPING, UNAVAILABLE, timeout, or validation failure is a hard stop; wait for OpenState before proposing a transition. Provider endpoints, credentials, aliases, catalogs, and raw provider errors are never returned to the client.`
	}
	srv := server.NewMCPServer("openstate", "0.1.0", server.WithInstructions(instructions))
	deps.GatewayMode = mode
	registerTools(srv, deps)
	return srv
}

// getArgs safely extracts the tool arguments as a map.
func getArgs(req mcp.CallToolRequest) map[string]any {
	args, _ := req.Params.Arguments.(map[string]any)
	if args == nil {
		return map[string]any{}
	}
	return args
}

// registerTools registers the standard OpenState MCP tools.
func registerTools(srv *server.MCPServer, deps Dependencies) {
	// list_intents
	listIntentsTool := mcp.NewTool("list_intents",
		mcp.WithDescription("List canonical intents and example user utterances for the authenticated tenant/project. Use the returned intent id with resolve_intent."),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithOpenWorldHintAnnotation(false),
		mcp.WithString("project", mcp.Description("Allowed project id; defaults to the API key default project")),
	)
	srv.AddTool(listIntentsTool, authorizedTool(deps, entities.MCPAPIScopeStateRead, func(ctx context.Context, principal entities.APIKeyPrincipal, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := getArgs(req)
		projectID, err := projectForPrincipal(ctx, deps, principal, str(args, "project"))
		if err != nil {
			return toolError(err)
		}
		return handleListIntents(ctx, deps, principal.TenantID, projectID)
	}))

	// resolve_intent
	resolveTool := mcp.NewTool("resolve_intent",
		mcp.WithDescription("Resolve a canonical intent id from list_intents to its mapped workflow. Do not pass a workflow id or slug."),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithOpenWorldHintAnnotation(false),
		mcp.WithString("intent", mcp.Required(), mcp.Description("Intent id, e.g. BOOKING_PADEL")),
		mcp.WithString("project", mcp.Description("Allowed project id; defaults to the API key default project")),
	)
	srv.AddTool(resolveTool, authorizedTool(deps, entities.MCPAPIScopeStateRead, func(ctx context.Context, principal entities.APIKeyPrincipal, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := getArgs(req)
		projectID, err := projectForPrincipal(ctx, deps, principal, str(args, "project"))
		if err != nil {
			return toolError(err)
		}
		return handleResolveIntent(ctx, deps, principal.TenantID, projectID, str(args, "intent"))
	}))

	// get_active_workflow
	activeTool := mcp.NewTool("get_active_workflow",
		mcp.WithDescription("Return the active workflow and current state for a conversation."),
		mcp.WithString("conversation", mcp.Required(), mcp.Description("Conversation id")),
	)
	srv.AddTool(activeTool, authorizedTool(deps, entities.MCPAPIScopeStateRead, func(ctx context.Context, principal entities.APIKeyPrincipal, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := getArgs(req)
		return handleGetActiveWorkflow(ctx, deps, principal.TenantID, str(args, "conversation"))
	}))

	// invoke_capability. Secure mode exposes only logical workflow context; the
	// server resolves the provider target internally from the active binding.
	if deps.GatewayMode == appservices.MCPGatewayModeSecure {
		invokeTool := mcp.NewTool("invoke_capability",
			mcp.WithDescription("Invoke a capability through the enforced OpenState gateway. OpenState derives the current state, project binding, provider connection, and exact tool; provider routing fields are not accepted."),
			mcp.WithString("instance", mcp.Required(), mcp.Description("Workflow instance id")),
			mcp.WithString("capability", mcp.Required(), mcp.Description("Logical capability declared by the current state")),
			mcp.WithString("correlationId", mcp.Required(), mcp.Description("Correlation identifier for this capability operation")),
			mcp.WithString("idempotencyKey", mcp.Required(), mcp.Description("Stable key preventing duplicate provider side effects")),
			mcp.WithObject("payload", mcp.Description("Capability input")),
		)
		srv.AddTool(invokeTool, authorizedTool(deps, entities.MCPAPIScopeCapabilityInvoke, func(ctx context.Context, principal entities.APIKeyPrincipal, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return handleGatewayInvokeCapability(ctx, deps, principal.TenantID, req)
		}))
	} else {
		invokeTool := mcp.NewTool("invoke_capability",
			mcp.WithDescription("Invoke an authorized capability for the context. Advisory mode does not prevent a client from connecting directly to a provider MCP."),
			mcp.WithString("workflow", mcp.Description("Workflow id")),
			mcp.WithString("state", mcp.Description("State id")),
			mcp.WithString("instance", mcp.Description("Workflow instance id — when provided, capability response is persisted to context")),
			mcp.WithString("capability", mcp.Required(), mcp.Description("Capability name, e.g. payment.create")),
			mcp.WithObject("payload", mcp.Description("Invocation payload")),
		)
		srv.AddTool(invokeTool, authorizedTool(deps, entities.MCPAPIScopeCapabilityInvoke, func(ctx context.Context, principal entities.APIKeyPrincipal, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return handleInvokeCapability(ctx, deps, principal.TenantID, req)
		}))
	}

	registerOrchestratorTools(srv, deps)
	registerContextTool(srv, deps)
}

// registerOrchestratorTools registers the runtime orchestration tools.
func registerOrchestratorTools(srv *server.MCPServer, deps Dependencies) {
	// get_current_state
	srv.AddTool(mcp.NewTool("get_current_state",
		mcp.WithDescription("Return the current state and allowed events for a workflow instance."),
		mcp.WithString("instance", mcp.Required(), mcp.Description("Workflow instance id")),
	), authorizedTool(deps, entities.MCPAPIScopeStateRead, func(ctx context.Context, principal entities.APIKeyPrincipal, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := getArgs(req)
		return handleGetCurrentState(ctx, deps, principal.TenantID, str(args, "instance"))
	}))

	// get_allowed_capabilities
	srv.AddTool(mcp.NewTool("get_allowed_capabilities",
		mcp.WithDescription("List capabilities authorized for a scope."),
		mcp.WithString("scopeType", mcp.Required(), mcp.Description("TENANT | WORKFLOW | STATE")),
		mcp.WithString("scopeId", mcp.Required(), mcp.Description("Scope id")),
		mcp.WithString("project", mcp.Description("Allowed project id; defaults to the API key default project")),
	), authorizedTool(deps, entities.MCPAPIScopeStateRead, func(ctx context.Context, principal entities.APIKeyPrincipal, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := getArgs(req)
		projectID, err := projectForPrincipal(ctx, deps, principal, str(args, "project"))
		if err != nil {
			return toolError(err)
		}
		return handleGetAllowedCapabilities(ctx, deps, principal.TenantID, projectID, str(args, "scopeType"), str(args, "scopeId"))
	}))

	// propose_event
	srv.AddTool(mcp.NewTool("propose_event",
		mcp.WithDescription("Propose an event for a workflow instance; the engine validates and transitions."),
		mcp.WithString("instance", mcp.Required(), mcp.Description("Workflow instance id")),
		mcp.WithString("type", mcp.Required(), mcp.Description("Event type")),
		mcp.WithObject("payload", mcp.Description("Event payload")),
	), authorizedTool(deps, entities.MCPAPIScopeStateWrite, func(ctx context.Context, principal entities.APIKeyPrincipal, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return handleProposeEvent(ctx, deps, principal.TenantID, req)
	}))

	// report_capability_result is retained only for advisory direct-two-MCP
	// compatibility. It is not part of the secure gateway surface.
	if deps.GatewayMode != appservices.MCPGatewayModeSecure {
		srv.AddTool(mcp.NewTool("report_capability_result",
			mcp.WithDescription("Report a direct provider MCP result to State MCP. A successful report is required before a guarded state transition can proceed."),
			mcp.WithString("project", mcp.Description("Allowed project id; defaults to the API key default project")),
			mcp.WithString("instance", mcp.Required(), mcp.Description("Workflow instance id")),
			mcp.WithString("state", mcp.Required(), mcp.Description("Current state id or state key")),
			mcp.WithString("capability", mcp.Required(), mcp.Description("Logical capability declared by the current state")),
			mcp.WithString("providerServer", mcp.Required(), mcp.Description("Configured provider MCP server alias; endpoints are not accepted")),
			mcp.WithString("providerTool", mcp.Required(), mcp.Description("Concrete provider MCP tool name")),
			mcp.WithString("correlationId", mcp.Required(), mcp.Description("Correlation identifier for the provider call")),
			mcp.WithString("idempotencyKey", mcp.Required(), mcp.Description("Stable key for the provider operation")),
			mcp.WithString("status", mcp.Required(), mcp.Description("SUCCEEDED or FAILED")),
			mcp.WithObject("result", mcp.Description("Normalized provider result")),
			mcp.WithObject("error", mcp.Description("Classified provider failure when status is FAILED")),
		), authorizedTool(deps, entities.MCPAPIScopeStateWrite, func(ctx context.Context, principal entities.APIKeyPrincipal, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			projectID, err := projectForPrincipal(ctx, deps, principal, str(getArgs(req), "project"))
			if err != nil {
				return toolError(err)
			}
			return handleReportCapabilityResult(ctx, deps, principal.TenantID, projectID, req)
		}))
	}

	// start_workflow
	srv.AddTool(mcp.NewTool("start_workflow",
		mcp.WithDescription("Start a new workflow instance."),
		mcp.WithString("workflow", mcp.Required(), mcp.Description("Workflow id")),
		mcp.WithString("version", mcp.Description("Workflow version id (optional)")),
		mcp.WithString("correlation", mcp.Description("Business/conversation correlation")),
	), authorizedTool(deps, entities.MCPAPIScopeStateWrite, func(ctx context.Context, principal entities.APIKeyPrincipal, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := getArgs(req)
		return handleStartWorkflow(ctx, deps, principal.TenantID, str(args, "workflow"), str(args, "version"), str(args, "correlation"))
	}))

	// suspend_workflow
	srv.AddTool(mcp.NewTool("suspend_workflow",
		mcp.WithDescription("Suspend a running workflow instance."),
		mcp.WithString("instance", mcp.Required(), mcp.Description("Workflow instance id")),
	), authorizedTool(deps, entities.MCPAPIScopeStateWrite, func(ctx context.Context, principal entities.APIKeyPrincipal, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := getArgs(req)
		return handleLifecycle(ctx, deps, "suspend", principal.TenantID, str(args, "instance"))
	}))

	// resume_workflow
	srv.AddTool(mcp.NewTool("resume_workflow",
		mcp.WithDescription("Resume a suspended workflow instance."),
		mcp.WithString("instance", mcp.Required(), mcp.Description("Workflow instance id")),
	), authorizedTool(deps, entities.MCPAPIScopeStateWrite, func(ctx context.Context, principal entities.APIKeyPrincipal, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := getArgs(req)
		return handleLifecycle(ctx, deps, "resume", principal.TenantID, str(args, "instance"))
	}))

	// cancel_workflow
	srv.AddTool(mcp.NewTool("cancel_workflow",
		mcp.WithDescription("Cancel a workflow instance."),
		mcp.WithString("instance", mcp.Required(), mcp.Description("Workflow instance id")),
	), authorizedTool(deps, entities.MCPAPIScopeStateWrite, func(ctx context.Context, principal entities.APIKeyPrincipal, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := getArgs(req)
		return handleLifecycle(ctx, deps, "cancel", principal.TenantID, str(args, "instance"))
	}))

	// get_workflow_instances
	srv.AddTool(mcp.NewTool("get_workflow_instances",
		mcp.WithDescription("List workflow instances for the authenticated tenant."),
	), authorizedTool(deps, entities.MCPAPIScopeStateRead, func(ctx context.Context, principal entities.APIKeyPrincipal, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return handleListInstances(ctx, deps, principal.TenantID)
	}))

	// get_history
	srv.AddTool(mcp.NewTool("get_history",
		mcp.WithDescription("Return the event history for a workflow instance."),
		mcp.WithString("instance", mcp.Required(), mcp.Description("Workflow instance id")),
	), authorizedTool(deps, entities.MCPAPIScopeStateRead, func(ctx context.Context, principal entities.APIKeyPrincipal, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := getArgs(req)
		return handleListHistory(ctx, deps, principal.TenantID, str(args, "instance"))
	}))

	// replay_workflow
	srv.AddTool(mcp.NewTool("replay_workflow",
		mcp.WithDescription("Replay the event history of a workflow instance to reproduce its resulting state."),
		mcp.WithString("instance", mcp.Required(), mcp.Description("Workflow instance id")),
	), authorizedTool(deps, entities.MCPAPIScopeStateRead, func(ctx context.Context, principal entities.APIKeyPrincipal, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := getArgs(req)
		return handleReplayWorkflow(ctx, deps, principal.TenantID, str(args, "instance"))
	}))
}

// registerContextTool registers the context compiler backed get_context tool.
func registerContextTool(srv *server.MCPServer, deps Dependencies) {
	srv.AddTool(mcp.NewTool("get_context",
		mcp.WithDescription("Return the compiled, PII-redacted runtime context for a turn."),
		mcp.WithString("conversation", mcp.Required(), mcp.Description("Conversation id")),
		mcp.WithString("instance", mcp.Description("Workflow instance id (optional)")),
		mcp.WithString("ownerType", mcp.Description("Memory owner type (e.g. CUSTOMER)")),
		mcp.WithString("ownerId", mcp.Description("Memory owner id")),
		mcp.WithString("query", mcp.Description("RAG query (optional)")),
	), authorizedTool(deps, entities.MCPAPIScopeStateRead, func(ctx context.Context, principal entities.APIKeyPrincipal, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := getArgs(req)
		return handleCompiledContext(ctx, deps, principal.TenantID, args)
	}))
}

// str returns a string argument or the empty string.
func str(args map[string]any, key string) string {
	s, _ := args[key].(string)
	return s
}
