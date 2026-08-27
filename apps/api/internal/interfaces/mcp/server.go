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

	"github.com/irosadie/open-state/api/internal/domain/capability"
)

// Dependencies bundles the domain services the MCP server needs. This keeps
// the package decoupled from concrete infrastructure wiring.
type Dependencies struct {
	// IntentResolver resolves an intent to its workflow + initial state.
	IntentResolver interface {
		ListIntents() []IntentInfo
	}
	// CapabilityInvoker executes authorized capabilities.
	CapabilityInvoker *capability.CapabilityInvoker
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
		mcp.WithDescription("Resolve a conversation intent to its workflow and current state."),
		mcp.WithString("intent", mcp.Required(), mcp.Description("Intent id, e.g. BOOKING_PADEL")),
		mcp.WithString("project", mcp.Required(), mcp.Description("Project id owning the intent")),
	)
	srv.AddTool(resolveTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := getArgs(req)
		intentID, _ := args["intent"].(string)
		projectID, _ := args["project"].(string)
		return handleResolveIntent(ctx, deps, intentID, projectID)
	})

	// get_active_workflow
	activeTool := mcp.NewTool("get_active_workflow",
		mcp.WithDescription("Return the active workflow and current state for a conversation."),
		mcp.WithString("conversation", mcp.Required(), mcp.Description("Conversation id")),
	)
	srv.AddTool(activeTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := getArgs(req)
		conv, _ := args["conversation"].(string)
		return handleGetActiveWorkflow(ctx, deps, conv)
	})

	// get_context
	ctxTool := mcp.NewTool("get_context",
		mcp.WithDescription("Return the compiled runtime context for a turn."),
		mcp.WithString("conversation", mcp.Required(), mcp.Description("Conversation id")),
		mcp.WithString("workflowInstance", mcp.Description("Workflow instance id (optional)")),
	)
	srv.AddTool(ctxTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := getArgs(req)
		conv, _ := args["conversation"].(string)
		wfi, _ := args["workflowInstance"].(string)
		return handleGetContext(ctx, deps, conv, wfi)
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
}
