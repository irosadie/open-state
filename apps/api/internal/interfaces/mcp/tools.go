package mcp

import (
	"context"
	"errors"

	"github.com/mark3labs/mcp-go/mcp"

	appservices "github.com/irosadie/open-state/api/internal/application/services"
	"github.com/irosadie/open-state/api/internal/domain/capability"
	"github.com/irosadie/open-state/api/internal/domain/entities"
)

// handleResolveIntent resolves an intent to its workflow (PRD 38, 171).
func handleResolveIntent(ctx context.Context, deps Dependencies, tenantID, projectID, intentID string) (*mcp.CallToolResult, error) {
	if deps.IntentResolver == nil {
		return toolUnavailable("intent resolver not configured")
	}
	wf, err := deps.IntentResolver.ResolveIntent(ctx, tenantID, projectID, intentID)
	if err != nil {
		return toolError(err)
	}
	return mcp.NewToolResultJSON(map[string]any{
		"intent":       intentID,
		"projectId":    projectID,
		"workflowId":   wf.ID,
		"workflowSlug": wf.Slug,
		"name":         wf.Name,
		"status":       wf.Status,
		"resolved":     true,
	})
}

// handleGetActiveWorkflow resolves the active workflow instance for a conversation
// (PRD 10, 142).
func handleGetActiveWorkflow(ctx context.Context, deps Dependencies, tenantID, conversation string) (*mcp.CallToolResult, error) {
	if deps.Orchestrator == nil {
		return toolUnavailable("orchestrator not configured")
	}
	inst, err := deps.Orchestrator.GetActiveWorkflow(ctx, tenantID, conversation)
	if err != nil {
		return toolError(err)
	}
	return mcp.NewToolResultJSON(map[string]any{
		"conversation": conversation,
		"active":       true,
		"instanceId":   inst.ID,
		"workflow":     inst.WorkflowID,
		"status":       inst.Status,
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

// ---- orchestrator tool handlers ----

func handleGetCurrentState(ctx context.Context, deps Dependencies, tenantID, instanceID string) (*mcp.CallToolResult, error) {
	if deps.Orchestrator == nil {
		return toolUnavailable("orchestrator not configured")
	}
	inst, stateInst, err := deps.Orchestrator.GetCurrentState(ctx, tenantID, instanceID)
	if err != nil {
		return toolError(err)
	}
	out := map[string]any{
		"instanceId": inst.ID,
		"status":     inst.Status,
	}
	if stateInst != nil {
		out["stateId"] = stateInst.ID
		out["stateKey"] = stateInst.StateKey
		out["stateStatus"] = stateInst.Status
	}

	// Allowed events/transitions from the current state (PRD 12, 14, 33-34).
	if transitions, terr := deps.Orchestrator.GetAllowedTransitions(ctx, tenantID, instanceID); terr == nil {
		list := make([]map[string]any, 0, len(transitions))
		for _, t := range transitions {
			list = append(list, map[string]any{
				"event":    t.Event,
				"targetStateId": t.TargetStateID,
				"priority": t.Priority,
			})
		}
		out["allowedTransitions"] = list
	}

	// Purpose/instructions/context of the current node (PRD 12, 14).
	if info, ierr := deps.Orchestrator.CurrentStateInfo(ctx, tenantID, instanceID); ierr == nil {
		out["purpose"] = info.Purpose
		out["instructions"] = info.Instructions
		out["requiredContext"] = info.RequiredContext
	}
	return mcp.NewToolResultJSON(out)
}

func handleGetAllowedCapabilities(ctx context.Context, deps Dependencies, tenantID, scopeType, scopeID string) (*mcp.CallToolResult, error) {
	if deps.Orchestrator == nil {
		return toolUnavailable("orchestrator not configured")
	}
	caps, err := deps.Orchestrator.ListAllowedCapabilities(ctx, tenantID, entities.BindingScopeType(scopeType), scopeID)
	if err != nil {
		return toolError(err)
	}
	out := make([]map[string]any, 0, len(caps))
	for _, c := range caps {
		out = append(out, map[string]any{
			"id":      c.ID,
			"name":    c.Name,
			"type":    c.ProviderType,
			"status":  c.Status,
		})
	}
	return mcp.NewToolResultJSON(map[string]any{"capabilities": out})
}

func handleProposeEvent(ctx context.Context, deps Dependencies, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if deps.Orchestrator == nil {
		return toolUnavailable("orchestrator not configured")
	}
	args := getArgs(req)
	tenantID := str(args, "tenant")
	instanceID := str(args, "instance")
	eventType := str(args, "type")
	payload := map[string]any{}
	if raw, ok := args["payload"]; ok {
		if m, ok := raw.(map[string]any); ok {
			payload = m
		}
	}
	evt, err := deps.Orchestrator.ProposeEvent(ctx, tenantID, instanceID, eventType, payload, "")
	if err != nil {
		return toolError(err)
	}
	return mcp.NewToolResultJSON(map[string]any{
		"ok":        true,
		"eventId":   evt.ID,
		"eventType": evt.Type,
		"sequence":  evt.Sequence,
	})
}

func handleStartWorkflow(ctx context.Context, deps Dependencies, tenantID, workflowID, versionID, correlation string) (*mcp.CallToolResult, error) {
	if deps.Orchestrator == nil {
		return toolUnavailable("orchestrator not configured")
	}
	inst, err := deps.Orchestrator.StartWorkflow(ctx, tenantID, workflowID, versionID, correlation)
	if err != nil {
		return toolError(err)
	}
	return mcp.NewToolResultJSON(map[string]any{
		"ok":         true,
		"instanceId": inst.ID,
		"status":     inst.Status,
	})
}

func handleLifecycle(ctx context.Context, deps Dependencies, action, tenantID, instanceID string) (*mcp.CallToolResult, error) {
	if deps.Orchestrator == nil {
		return toolUnavailable("orchestrator not configured")
	}
	var (
		inst *entities.WorkflowInstance
		err  error
	)
	switch action {
	case "suspend":
		inst, err = deps.Orchestrator.SuspendWorkflow(ctx, tenantID, instanceID)
	case "resume":
		inst, err = deps.Orchestrator.ResumeWorkflow(ctx, tenantID, instanceID)
	case "cancel":
		inst, err = deps.Orchestrator.CancelWorkflow(ctx, tenantID, instanceID)
	default:
		return toolUnavailable("unknown lifecycle action")
	}
	if err != nil {
		return toolError(err)
	}
	return mcp.NewToolResultJSON(map[string]any{
		"ok":         true,
		"instanceId": inst.ID,
		"status":     inst.Status,
	})
}

func handleListInstances(ctx context.Context, deps Dependencies, tenantID string) (*mcp.CallToolResult, error) {
	if deps.Orchestrator == nil {
		return toolUnavailable("orchestrator not configured")
	}
	insts, err := deps.Orchestrator.ListInstances(ctx, tenantID)
	if err != nil {
		return toolError(err)
	}
	out := make([]map[string]any, 0, len(insts))
	for _, i := range insts {
		out = append(out, map[string]any{
			"id":     i.ID,
			"workflow": i.WorkflowID,
			"status": i.Status,
			"version": i.Version,
		})
	}
	return mcp.NewToolResultJSON(map[string]any{"instances": out})
}

func handleListHistory(ctx context.Context, deps Dependencies, tenantID, instanceID string) (*mcp.CallToolResult, error) {
	if deps.Orchestrator == nil {
		return toolUnavailable("orchestrator not configured")
	}
	events, err := deps.Orchestrator.ListHistory(ctx, tenantID, instanceID)
	if err != nil {
		return toolError(err)
	}
	out := make([]map[string]any, 0, len(events))
	for _, e := range events {
		out = append(out, map[string]any{
			"id":       e.ID,
			"type":     e.Type,
			"sequence": e.Sequence,
			"source":   e.Source,
			"timestamp": e.Timestamp,
		})
	}
	return mcp.NewToolResultJSON(map[string]any{"history": out})
}

func handleReplayWorkflow(ctx context.Context, deps Dependencies, tenantID, instanceID string) (*mcp.CallToolResult, error) {
	if deps.Orchestrator == nil {
		return toolUnavailable("orchestrator not configured")
	}
	contextSnap, stateKey, err := deps.Orchestrator.ReplayState(ctx, tenantID, instanceID)
	if err != nil {
		return toolError(err)
	}
	out := map[string]any{
		"replayed":   true,
		"context":    contextSnap,
		"stateKey":   stateKey,
	}
	return mcp.NewToolResultJSON(out)
}

func handleCompiledContext(ctx context.Context, deps Dependencies, args map[string]any) (*mcp.CallToolResult, error) {
	if deps.ContextCompiler == nil {
		return toolUnavailable("context compiler not configured")
	}
	compiled, err := deps.ContextCompiler.Compile(ctx, appservices.CompileArgs{
		TenantID:       str(args, "tenant"),
		ConversationID: str(args, "conversation"),
		WorkflowInstanceID: str(args, "instance"),
		OwnerType:      str(args, "ownerType"),
		OwnerID:        str(args, "ownerId"),
		Query:          str(args, "query"),
	})
	if err != nil {
		return toolError(err)
	}
	return mcp.NewToolResultJSON(map[string]any{
		"available": compiled.Available,
		"missing":   compiled.Missing,
		"memory":    compiled.Memory,
		"workflow":  compiled.Workflow,
		"retrieval": compiled.Retrieval,
		"redacted":  compiled.Redacted,
	})
}

// ---- shared helpers ----

func toolUnavailable(msg string) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultJSON(map[string]any{"ok": false, "message": msg})
}

func toolError(err error) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultJSON(map[string]any{"ok": false, "message": err.Error()})
}
