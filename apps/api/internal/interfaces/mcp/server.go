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
)

// Dependencies bundles the domain services the MCP server needs. This keeps
// the package decoupled from concrete infrastructure wiring.
type Dependencies struct {
	// IntentResolver resolves an intent to its workflow (PRD 38, 171).
	IntentResolver IntentPort
	// CapabilityInvoker executes authorized capabilities.
	CapabilityInvoker *capability.CapabilityInvoker
	// Orchestrator exposes runtime workflow orchestration (lifecycle, history,
	// event proposal, instances, allowed capabilities).
	Orchestrator OrchestratorPort
	// ContextCompiler compiles minimal per-turn context (PRD 22).
	ContextCompiler ContextCompilerPort
}

// IntentPort resolves conversation intents to workflows.
type IntentPort interface {
	ListIntents(ctx context.Context, tenantID, projectID string) ([]entities.Workflow, error)
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
	ListAllowedCapabilities(ctx context.Context, tenantID string, scopeType entities.BindingScopeType, scopeID string) ([]entities.Capability, error)
}

// ContextCompilerPort is the application-layer seam the get_context tool delegates to.
type ContextCompilerPort interface {
	Compile(ctx context.Context, args appservices.CompileArgs) (*dtos.CompiledContext, error)
}

// IntentInfo is a lightweight projection of an intent for the MCP tool.
type IntentInfo struct {
	ID           string `json:"id"`
	ProjectID    string `json:"projectId"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	WorkflowSlug string `json:"workflowSlug"`
}

// NewServer builds an MCP server with the standard OpenState toolset.
func NewServer(deps Dependencies) *server.MCPServer {
	srv := server.NewMCPServer("openstate", "0.1.0")
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

// registerTools registers the four MCP tools.
func registerTools(srv *server.MCPServer, deps Dependencies) {
	// resolve_intent
	resolveTool := mcp.NewTool("resolve_intent",
		mcp.WithDescription("Resolve a conversation intent to its workflow."),
		mcp.WithString("tenant", mcp.Required(), mcp.Description("Tenant id")),
		mcp.WithString("intent", mcp.Required(), mcp.Description("Intent id, e.g. BOOKING_PADEL")),
		mcp.WithString("project", mcp.Required(), mcp.Description("Project id owning the intent")),
	)
	srv.AddTool(resolveTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := getArgs(req)
		intentID, _ := args["intent"].(string)
		projectID, _ := args["project"].(string)
		tenantID, _ := args["tenant"].(string)
		return handleResolveIntent(ctx, deps, tenantID, projectID, intentID)
	})

	// get_active_workflow
	activeTool := mcp.NewTool("get_active_workflow",
		mcp.WithDescription("Return the active workflow and current state for a conversation."),
		mcp.WithString("tenant", mcp.Required(), mcp.Description("Tenant id")),
		mcp.WithString("conversation", mcp.Required(), mcp.Description("Conversation id")),
	)
	srv.AddTool(activeTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := getArgs(req)
		conv, _ := args["conversation"].(string)
		tenantID, _ := args["tenant"].(string)
		return handleGetActiveWorkflow(ctx, deps, tenantID, conv)
	})

	// invoke_capability
	invokeTool := mcp.NewTool("invoke_capability",
		mcp.WithDescription("Invoke an authorized capability for the context."),
		mcp.WithString("tenant", mcp.Required(), mcp.Description("Tenant id")),
		mcp.WithString("workflow", mcp.Description("Workflow id")),
		mcp.WithString("state", mcp.Description("State id")),
		mcp.WithString("capability", mcp.Required(), mcp.Description("Capability name, e.g. payment.create")),
		mcp.WithObject("payload", mcp.Description("Invocation payload")),
	)
	srv.AddTool(invokeTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return handleInvokeCapability(ctx, deps, req)
	})

	registerOrchestratorTools(srv, deps)
	registerContextTool(srv, deps)
}

// registerOrchestratorTools registers the runtime orchestration tools.
func registerOrchestratorTools(srv *server.MCPServer, deps Dependencies) {
	// get_current_state
	srv.AddTool(mcp.NewTool("get_current_state",
		mcp.WithDescription("Return the current state and allowed events for a workflow instance."),
		mcp.WithString("tenant", mcp.Required(), mcp.Description("Tenant id")),
		mcp.WithString("instance", mcp.Required(), mcp.Description("Workflow instance id")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := getArgs(req)
		return handleGetCurrentState(ctx, deps, str(args, "tenant"), str(args, "instance"))
	})

	// get_allowed_capabilities
	srv.AddTool(mcp.NewTool("get_allowed_capabilities",
		mcp.WithDescription("List capabilities authorized for a scope."),
		mcp.WithString("tenant", mcp.Required(), mcp.Description("Tenant id")),
		mcp.WithString("scopeType", mcp.Required(), mcp.Description("TENANT | WORKFLOW | STATE")),
		mcp.WithString("scopeId", mcp.Required(), mcp.Description("Scope id")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := getArgs(req)
		return handleGetAllowedCapabilities(ctx, deps, str(args, "tenant"), str(args, "scopeType"), str(args, "scopeId"))
	})

	// propose_event
	srv.AddTool(mcp.NewTool("propose_event",
		mcp.WithDescription("Propose an event for a workflow instance; the engine validates and transitions."),
		mcp.WithString("tenant", mcp.Required(), mcp.Description("Tenant id")),
		mcp.WithString("instance", mcp.Required(), mcp.Description("Workflow instance id")),
		mcp.WithString("type", mcp.Required(), mcp.Description("Event type")),
		mcp.WithObject("payload", mcp.Description("Event payload")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return handleProposeEvent(ctx, deps, req)
	})

	// start_workflow
	srv.AddTool(mcp.NewTool("start_workflow",
		mcp.WithDescription("Start a new workflow instance."),
		mcp.WithString("tenant", mcp.Required(), mcp.Description("Tenant id")),
		mcp.WithString("workflow", mcp.Required(), mcp.Description("Workflow id")),
		mcp.WithString("version", mcp.Description("Workflow version id (optional)")),
		mcp.WithString("correlation", mcp.Description("Business/conversation correlation")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := getArgs(req)
		return handleStartWorkflow(ctx, deps, str(args, "tenant"), str(args, "workflow"), str(args, "version"), str(args, "correlation"))
	})

	// suspend_workflow
	srv.AddTool(mcp.NewTool("suspend_workflow",
		mcp.WithDescription("Suspend a running workflow instance."),
		mcp.WithString("tenant", mcp.Required(), mcp.Description("Tenant id")),
		mcp.WithString("instance", mcp.Required(), mcp.Description("Workflow instance id")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := getArgs(req)
		return handleLifecycle(ctx, deps, "suspend", str(args, "tenant"), str(args, "instance"))
	})

	// resume_workflow
	srv.AddTool(mcp.NewTool("resume_workflow",
		mcp.WithDescription("Resume a suspended workflow instance."),
		mcp.WithString("tenant", mcp.Required(), mcp.Description("Tenant id")),
		mcp.WithString("instance", mcp.Required(), mcp.Description("Workflow instance id")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := getArgs(req)
		return handleLifecycle(ctx, deps, "resume", str(args, "tenant"), str(args, "instance"))
	})

	// cancel_workflow
	srv.AddTool(mcp.NewTool("cancel_workflow",
		mcp.WithDescription("Cancel a workflow instance."),
		mcp.WithString("tenant", mcp.Required(), mcp.Description("Tenant id")),
		mcp.WithString("instance", mcp.Required(), mcp.Description("Workflow instance id")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := getArgs(req)
		return handleLifecycle(ctx, deps, "cancel", str(args, "tenant"), str(args, "instance"))
	})

	// get_workflow_instances
	srv.AddTool(mcp.NewTool("get_workflow_instances",
		mcp.WithDescription("List workflow instances for a tenant."),
		mcp.WithString("tenant", mcp.Required(), mcp.Description("Tenant id")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := getArgs(req)
		return handleListInstances(ctx, deps, str(args, "tenant"))
	})

	// get_history
	srv.AddTool(mcp.NewTool("get_history",
		mcp.WithDescription("Return the event history for a workflow instance."),
		mcp.WithString("tenant", mcp.Required(), mcp.Description("Tenant id")),
		mcp.WithString("instance", mcp.Required(), mcp.Description("Workflow instance id")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := getArgs(req)
		return handleListHistory(ctx, deps, str(args, "tenant"), str(args, "instance"))
	})

	// replay_workflow
	srv.AddTool(mcp.NewTool("replay_workflow",
		mcp.WithDescription("Replay the event history of a workflow instance to reproduce its resulting state."),
		mcp.WithString("tenant", mcp.Required(), mcp.Description("Tenant id")),
		mcp.WithString("instance", mcp.Required(), mcp.Description("Workflow instance id")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := getArgs(req)
		return handleReplayWorkflow(ctx, deps, str(args, "tenant"), str(args, "instance"))
	})
}

// registerContextTool registers the context compiler backed get_context tool.
func registerContextTool(srv *server.MCPServer, deps Dependencies) {
	srv.AddTool(mcp.NewTool("get_context",
		mcp.WithDescription("Return the compiled, PII-redacted runtime context for a turn."),
		mcp.WithString("tenant", mcp.Required(), mcp.Description("Tenant id")),
		mcp.WithString("conversation", mcp.Required(), mcp.Description("Conversation id")),
		mcp.WithString("instance", mcp.Description("Workflow instance id (optional)")),
		mcp.WithString("ownerType", mcp.Description("Memory owner type (e.g. CUSTOMER)")),
		mcp.WithString("ownerId", mcp.Description("Memory owner id")),
		mcp.WithString("query", mcp.Description("RAG query (optional)")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := getArgs(req)
		return handleCompiledContext(ctx, deps, args)
	})
}

// str returns a string argument or the empty string.
func str(args map[string]any, key string) string {
	s, _ := args[key].(string)
	return s
}
