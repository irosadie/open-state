package mcp

import (
	"context"
	"errors"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/irosadie/open-state/api/internal/domain/capability"
)

// handleResolveIntent returns the resolved intent's workflow + state projection.
func handleResolveIntent(_ context.Context, deps Dependencies, intentID, projectID string) (*mcp.CallToolResult, error) {
	for _, i := range deps.IntentResolver.ListIntents() {
		if i.ID == intentID && (projectID == "" || i.ProjectID == projectID) {
			return mcp.NewToolResultJSON(map[string]any{
				"intent":       i.ID,
				"projectId":    i.ProjectID,
				"workflowSlug": i.WorkflowSlug,
				"resolved":     true,
			})
		}
	}
	return mcp.NewToolResultJSON(map[string]any{
		"intent":   intentID,
		"resolved": false,
		"reason":   "intent not found",
	})
}

// handleGetActiveWorkflow returns an indication of active workflow. In this
// slice the active-workflow store is not yet wired; it reports none found.
func handleGetActiveWorkflow(_ context.Context, _ Dependencies, conversation string) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultJSON(map[string]any{
		"conversation": conversation,
		"active":       false,
		"message":      "no active workflow (active-workflow store not wired in this slice)",
	})
}

// handleGetContext returns a stub compiled context. Full context resolution
// lands once engine-context-resolver is wired into the server.
func handleGetContext(_ context.Context, _ Dependencies, conversation, workflowInstance string) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultJSON(map[string]any{
		"conversation":     conversation,
		"workflowInstance": workflowInstance,
		"available":        map[string]any{},
		"missing":          []string{},
		"note":             "context resolver not wired in this slice",
	})
}

// handleInvokeCapability runs an authorized capability through the invoker.
func handleInvokeCapability(ctx context.Context, deps Dependencies, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := getArgs(req)
	tenant, _ := args["tenant"].(string)
	workflow, _ := args["workflow"].(string)
	state, _ := args["state"].(string)
	name, _ := args["capability"].(string)

	if deps.CapabilityInvoker == nil {
		return mcp.NewToolResultJSON(map[string]any{
			"ok":      false,
			"error":   "capability invoker not configured",
			"invoked": false,
		})
	}

	payload := map[string]any{}
	if raw, ok := args["payload"]; ok {
		if m, ok := raw.(map[string]any); ok {
			payload = m
		}
	}

	inv := capability.Invocation{
		TenantID:   tenant,
		WorkflowID: workflow,
		StateID:    state,
		Name:       name,
		Payload:    payload,
		Policy:     capability.CapabilityPolicy{Timeout: 10_000_000_000, MaxRetry: 0},
	}
	res, err := deps.CapabilityInvoker.Execute(ctx, inv)

	var ce *capability.CapabilityError
	if errors.As(err, &ce) {
		return mcp.NewToolResultJSON(map[string]any{
			"ok":      false,
			"kind":    ce.Kind,
			"code":    ce.Code,
			"message": ce.Message,
			"invoked": false,
		})
	}
	if err != nil {
		return mcp.NewToolResultJSON(map[string]any{
			"ok":      false,
			"kind":    "EXTERNAL",
			"message": err.Error(),
			"invoked": false,
		})
	}
	return mcp.NewToolResultJSON(map[string]any{
		"ok":             true,
		"data":           res.Data,
		"fromMock":       res.FromMock,
		"capabilityEvent": res.CapabilityEvent,
		"invoked":        true,
	})
}
