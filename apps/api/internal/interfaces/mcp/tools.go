package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	appservices "github.com/irosadie/open-state/api/internal/application/services"
	"github.com/irosadie/open-state/api/internal/domain/capability"
	"github.com/irosadie/open-state/api/internal/domain/engine"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
)

// handleResolveIntent resolves an intent to its workflow (PRD 38, 171).
func handleResolveIntent(ctx context.Context, deps Dependencies, tenantID, projectID, intentID string) (*mcp.CallToolResult, error) {
	if deps.IntentResolver == nil {
		return toolUnavailable("intent resolver not configured")
	}
	key := strings.ToUpper(strings.TrimSpace(intentID))
	if strings.TrimSpace(tenantID) == "" || strings.TrimSpace(projectID) == "" {
		return toolError(errors.New("tenant and project are required"))
	}
	if key == "" {
		return toolError(errors.New("intent is required"))
	}
	wf, err := deps.IntentResolver.ResolveIntent(ctx, tenantID, projectID, key)
	if err != nil {
		return toolError(err)
	}
	requirements, reqErr := entryRequirements(ctx, deps, tenantID, projectID, wf)
	if reqErr != nil {
		return toolError(reqErr)
	}
	return mcp.NewToolResultJSON(map[string]any{
		"intent":       key,
		"projectId":    projectID,
		"workflowId":   wf.ID,
		"workflowSlug": wf.Slug,
		"name":         wf.Name,
		"status":       wf.Status,
		"stateController": map[string]any{
			"server":               "openstate",
			"mandatory":            true,
			"readBeforeTransition": true,
		},
		"requiredCapabilities": requirements,
		"resolved":             true,
	})
}

// handleListIntents returns the canonical intent catalog for a project.
func handleListIntents(ctx context.Context, deps Dependencies, tenantID, projectID string) (*mcp.CallToolResult, error) {
	if deps.IntentResolver == nil {
		return toolUnavailable("intent resolver not configured")
	}
	if strings.TrimSpace(tenantID) == "" || strings.TrimSpace(projectID) == "" {
		return toolError(errors.New("tenant and project are required"))
	}
	intents, err := deps.IntentResolver.ListIntents(ctx, tenantID, projectID)
	if err != nil {
		return toolError(err)
	}
	out := make([]IntentInfo, 0, len(intents))
	for _, intent := range intents {
		out = append(out, IntentInfo{
			ID:           intent.Key,
			ProjectID:    intent.ProjectID,
			Name:         intent.Name,
			Description:  intent.Description,
			Examples:     intent.Examples,
			WorkflowSlug: intent.WorkflowSlug,
		})
	}
	return mcp.NewToolResultJSON(map[string]any{
		"tenant":    tenantID,
		"projectId": projectID,
		"intents":   out,
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
	recordRuntimeTrace(ctx, deps, tenantID, inst.ID, appservices.TraceRecordInput{
		Stage:         entities.RuntimeTraceStageWorkflowLookup,
		Status:        entities.RuntimeTraceStatusSucceeded,
		CorrelationID: traceStringPtr(conversation),
		Attributes: map[string]any{
			"workflow_id": inst.WorkflowID,
			"status":      inst.Status,
		},
	})
	return mcp.NewToolResultJSON(map[string]any{
		"conversation": conversation,
		"active":       true,
		"instanceId":   inst.ID,
		"workflow":     inst.WorkflowID,
		"status":       inst.Status,
	})
}

// handleInvokeCapability runs an authorized capability through the invoker
// and persists the response data into workflow instance context so downstream
// capabilities can read it without re-invoking (PRD §24, §31).
func handleInvokeCapability(ctx context.Context, deps Dependencies, tenant string, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := getArgs(req)
	workflow, _ := args["workflow"].(string)
	state, _ := args["state"].(string)
	name, _ := args["capability"].(string)
	instance, _ := args["instance"].(string)

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

	// Persist each top-level key from the capability response into the
	// workflow instance context so downstream capabilities and the context
	// compiler can read it without re-invoking (PRD §24, §31).
	// Only persisted when instance id is provided and ContextRepo is wired.
	if instance != "" && deps.ContextRepo != nil && len(res.Data) > 0 {
		persistCapabilityContext(ctx, deps, tenant, instance, name, res.Data)
	}

	return mcp.NewToolResultJSON(map[string]any{
		"ok":              true,
		"data":            res.Data,
		"fromMock":        res.FromMock,
		"capabilityEvent": res.CapabilityEvent,
		"invoked":         true,
	})
}

// handleGatewayInvokeCapability is the secure State MCP path. It accepts only
// workflow context and logical capability input; provider routing is resolved
// by MCPGatewayService from the current state and project binding.
func handleGatewayInvokeCapability(ctx context.Context, deps Dependencies, tenant string, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if deps.Gateway == nil {
		return mcp.NewToolResultJSON(secureGatewayFailure(
			capability.ErrorKindUnavailable,
			"capability.gateway_unavailable",
			"enforced MCP gateway is not configured",
		))
	}
	args := getArgs(req)
	payload := map[string]any{}
	if raw, ok := args["payload"].(map[string]any); ok && raw != nil {
		payload = raw
	}
	result, err := deps.Gateway.Execute(ctx, appservices.GatewayInvocationRequest{
		TenantID:       tenant,
		InstanceID:     strings.TrimSpace(str(args, "instance")),
		CapabilityName: strings.TrimSpace(str(args, "capability")),
		Payload:        payload,
		CorrelationID:  strings.TrimSpace(str(args, "correlationId")),
		IdempotencyKey: strings.TrimSpace(str(args, "idempotencyKey")),
	})
	if err != nil {
		var ce *capability.CapabilityError
		if errors.As(err, &ce) {
			return mcp.NewToolResultJSON(secureGatewayFailure(ce.Kind, ce.Code, ce.Message))
		}
		return mcp.NewToolResultJSON(secureGatewayFailure(
			capability.ErrorKindUnavailable,
			"capability.gateway_unavailable",
			"capability gateway is unavailable",
		))
	}
	return mcp.NewToolResultJSON(map[string]any{
		"ok": true, "invoked": true, "instanceId": result.InstanceID, "stateId": result.StateID,
		"capability": result.CapabilityName, "status": result.Status, "reused": result.Reused, "data": result.Data,
		"nextAction": "CONTINUE",
	})
}

// secureGatewayFailure is deliberately prescriptive. An LLM must be able to
// distinguish a blocked gateway call from a successful provider result without
// interpreting a safe error message as permission to search another scope.
func secureGatewayFailure(kind capability.ErrorKind, code, message string) map[string]any {
	return map[string]any{
		"ok":         false,
		"invoked":    false,
		"hardStop":   true,
		"nextAction": "STOP",
		"kind":       kind,
		"code":       code,
		"message":    message,
	}
}

func secureMutationFailure(code, message string) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultJSON(map[string]any{
		"ok":         false,
		"invoked":    false,
		"hardStop":   true,
		"nextAction": "STOP",
		"kind":       capability.ErrorKindValidation,
		"code":       code,
		"message":    message,
	})
}

// persistCapabilityContext writes each key in data to context_records scoped
// to the workflow instance. Keys are prefixed with the capability name to avoid
// collisions (e.g. "booking.confirm.booking_id"). Internal _-prefixed keys
// (echo, _input, _capability) are skipped.
func persistCapabilityContext(ctx context.Context, deps Dependencies, tenantID, instanceID, capabilityName string, data map[string]any) {
	for key, val := range data {
		// skip internal echo/meta keys added by JSONFileProvider
		if len(key) > 0 && key[0] == '_' {
			continue
		}
		raw, err := json.Marshal(val)
		if err != nil {
			continue
		}
		contextKey := capabilityName + "." + key
		// version 0 = insert-or-update without strict optimistic lock
		_, _ = deps.ContextRepo.UpsertContext(ctx, tenantID,
			entities.ContextScopeWorkflowInstance, instanceID,
			contextKey, raw, 0)
	}
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
	recordRuntimeTrace(ctx, deps, tenantID, instanceID, appservices.TraceRecordInput{
		Stage:  entities.RuntimeTraceStageStateLookup,
		Status: entities.RuntimeTraceStatusSucceeded,
		Attributes: map[string]any{
			"state_instance_id": stateInstanceID(stateInst),
			"instance_status":   inst.Status,
		},
	})
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
				"event":         t.Event,
				"targetStateId": t.TargetStateID,
				"priority":      t.Priority,
			})
		}
		out["allowedTransitions"] = list
	}

	// Purpose/instructions/context of the current node (PRD 12, 14).
	if info, ierr := deps.Orchestrator.CurrentStateInfo(ctx, tenantID, instanceID); ierr == nil {
		out["stateId"] = info.StateID
		projectID := info.ProjectID
		if projectID == "" && deps.WorkflowRegistry != nil {
			if version, versionErr := deps.WorkflowRegistry.FindCurrentVersionByWorkflow(ctx, tenantID, inst.WorkflowID); versionErr == nil {
				projectID = version.ProjectID
			}
		}
		out["projectId"] = projectID
		out["purpose"] = info.Purpose
		out["instructions"] = info.Instructions
		out["requiredContext"] = info.RequiredContext
		out["capabilities"] = info.Capabilities
		evidence := []entities.CapabilityExecutionEvidence{}
		if deps.CapabilityEvidence != nil {
			if found, eerr := deps.CapabilityEvidence.ListByInstanceState(ctx, tenantID, instanceID, info.StateID); eerr == nil {
				evidence = found
			}
		}
		requirements, rerr := requirementsForCapabilities(ctx, deps, tenantID, projectID, info.Capabilities, transitionEvents(out["allowedTransitions"]), evidence)
		if rerr != nil {
			return toolError(rerr)
		}
		out["requiredCapabilities"] = requirements
	}
	return mcp.NewToolResultJSON(out)
}

func handleGetAllowedCapabilities(ctx context.Context, deps Dependencies, tenantID, projectID, scopeType, scopeID string) (*mcp.CallToolResult, error) {
	if deps.Orchestrator == nil {
		return toolUnavailable("orchestrator not configured")
	}
	caps, err := deps.Orchestrator.ListAllowedCapabilities(ctx, tenantID, entities.BindingScopeType(scopeType), scopeID)
	if err != nil {
		return toolError(err)
	}
	out := make([]map[string]any, 0, len(caps))
	for _, c := range caps {
		requirements, _ := requirementsForCapabilities(ctx, deps, tenantID, projectID, []string{c.Name}, nil, nil)
		item := map[string]any{
			"id":     c.ID,
			"name":   c.Name,
			"type":   c.ProviderType,
			"status": c.Status,
		}
		if len(requirements) > 0 {
			item["provider"] = requirements[0]
		}
		out = append(out, item)
	}
	return mcp.NewToolResultJSON(map[string]any{"capabilities": out})
}

func transitionEvents(raw any) []string {
	items, ok := raw.([]map[string]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		if event, ok := item["event"].(string); ok && event != "" {
			out = append(out, event)
		}
	}
	return out
}

func handleReportCapabilityResult(ctx context.Context, deps Dependencies, tenantID, projectID string, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if deps.Orchestrator == nil || deps.CapabilityRegistry == nil {
		return toolUnavailable("state capability evidence dependencies are not configured")
	}
	if deps.CapabilityEvidence == nil {
		return toolUnavailable("capability evidence store is not configured")
	}
	args := getArgs(req)
	instanceID := strings.TrimSpace(str(args, "instance"))
	stateID := strings.TrimSpace(str(args, "state"))
	name := strings.TrimSpace(str(args, "capability"))
	providerServer := strings.TrimSpace(str(args, "providerServer"))
	providerTool := strings.TrimSpace(str(args, "providerTool"))
	correlationID := strings.TrimSpace(str(args, "correlationId"))
	idempotencyKey := strings.TrimSpace(str(args, "idempotencyKey"))
	status := strings.ToUpper(strings.TrimSpace(str(args, "status")))
	if instanceID == "" || stateID == "" || name == "" || providerServer == "" || providerTool == "" || correlationID == "" || idempotencyKey == "" {
		return toolError(errors.New("instance, state, capability, providerServer, providerTool, correlationId, and idempotencyKey are required"))
	}
	if strings.Contains(providerServer, "://") || strings.Contains(providerTool, "://") {
		return toolError(errors.New("provider endpoint is not accepted; use the configured provider server alias"))
	}
	if status != string(entities.CapabilityEvidenceSucceeded) && status != string(entities.CapabilityEvidenceFailed) {
		return toolError(errors.New("status must be SUCCEEDED or FAILED"))
	}
	inst, stateInst, err := deps.Orchestrator.GetCurrentState(ctx, tenantID, instanceID)
	if err != nil {
		return toolError(err)
	}
	info, err := deps.Orchestrator.CurrentStateInfo(ctx, tenantID, instanceID)
	if err != nil {
		return toolError(err)
	}
	currentProjectID := info.ProjectID
	if currentProjectID == "" && deps.WorkflowRegistry != nil {
		if version, versionErr := deps.WorkflowRegistry.FindCurrentVersionByWorkflow(ctx, tenantID, inst.WorkflowID); versionErr == nil {
			currentProjectID = version.ProjectID
		}
	}
	if currentProjectID != "" {
		if projectID != "" && projectID != currentProjectID {
			return toolError(errors.New("provider result project does not match the current workflow project"))
		}
		projectID = currentProjectID
	}
	if stateID != info.StateID && (stateInst == nil || stateID != stateInst.StateKey) && (stateInst == nil || stateID != stateInst.ID) {
		return toolError(errors.New("provider result state does not match the current state"))
	}
	declared := false
	for _, required := range info.Capabilities {
		if required == name {
			declared = true
			break
		}
	}
	if !declared {
		return toolError(errors.New("capability is not declared by the current state"))
	}
	cap, err := deps.CapabilityRegistry.FindByName(ctx, tenantID, name)
	if err != nil {
		return toolError(err)
	}
	if deps.ProjectCapabilityBindings == nil {
		return toolError(errors.New("project MCP capability binding resolver is not configured"))
	}
	binding, err := deps.ProjectCapabilityBindings.FindByCapability(ctx, tenantID, projectID, cap.ID)
	if err != nil {
		return toolError(errors.New("capability has no project MCP tool binding"))
	}
	if binding.Health != entities.ProjectCapabilityMCPBindingActive {
		return toolError(errors.New("capability provider binding is unavailable: " + binding.HealthReason))
	}
	if binding.ConnectionAlias != providerServer || binding.ToolName != providerTool {
		return toolError(errors.New("provider server or tool does not match the active project binding"))
	}
	if previous, findErr := deps.CapabilityEvidence.FindByIdempotency(ctx, tenantID, projectID, inst.ID, info.StateID, cap.ID, idempotencyKey); findErr == nil && previous.Status == entities.CapabilityEvidenceSucceeded {
		return mcp.NewToolResultJSON(map[string]any{
			"ok": true, "accepted": true, "reused": true,
			"instanceId": inst.ID, "stateId": info.StateID, "capability": cap.Name,
			"providerServer": previous.ProviderServer, "providerTool": previous.ProviderTool,
			"status": previous.Status, "correlationId": previous.CorrelationID, "idempotencyKey": previous.IdempotencyKey,
			"result": previous.Result,
		})
	}
	result := map[string]any{}
	if raw, ok := args["result"].(map[string]any); ok && raw != nil {
		result = raw
	}
	errorPayload := []byte(nil)
	if raw, ok := args["error"].(map[string]any); ok && raw != nil {
		errorPayload, err = json.Marshal(raw)
		if err != nil {
			return toolError(errors.New("invalid provider error payload"))
		}
	}
	if status == string(entities.CapabilityEvidenceFailed) && len(errorPayload) == 0 {
		return toolError(errors.New("error is required for a failed provider result"))
	}
	if status == string(entities.CapabilityEvidenceSucceeded) && deps.CapabilityOutputValidator != nil && len(cap.OutputSchema) > 0 {
		if err := deps.CapabilityOutputValidator.Validate(result, cap.OutputSchema); err != nil {
			return toolError(errors.New("provider result does not match output schema: " + err.Error()))
		}
	}
	rawResult, err := json.Marshal(result)
	if err != nil {
		return toolError(errors.New("invalid provider result payload"))
	}
	evidence, err := deps.CapabilityEvidence.Upsert(ctx, repositories.CapabilityEvidenceInput{
		TenantID: tenantID, ProjectID: projectID, WorkflowInstanceID: inst.ID, StateID: info.StateID,
		CapabilityID: cap.ID, CapabilityName: cap.Name, ProviderServer: providerServer, ProviderTool: providerTool,
		CorrelationID: &correlationID, IdempotencyKey: idempotencyKey, Status: entities.CapabilityEvidenceStatus(status),
		Result: rawResult, Error: errorPayload,
	})
	if err != nil {
		return toolError(err)
	}
	if status == string(entities.CapabilityEvidenceSucceeded) && deps.ContextRepo != nil && len(result) > 0 {
		persistCapabilityContext(ctx, deps, tenantID, instanceID, name, result)
	}
	return mcp.NewToolResultJSON(map[string]any{
		"ok": true, "accepted": status == string(entities.CapabilityEvidenceSucceeded),
		"instanceId": inst.ID, "stateId": info.StateID, "capability": cap.Name,
		"providerServer": evidence.ProviderServer, "providerTool": evidence.ProviderTool,
		"status": evidence.Status, "correlationId": evidence.CorrelationID, "idempotencyKey": evidence.IdempotencyKey,
	})
}

func handleProposeEvent(ctx context.Context, deps Dependencies, tenantID string, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if deps.Orchestrator == nil {
		return toolUnavailable("orchestrator not configured")
	}
	args := getArgs(req)
	instanceID := str(args, "instance")
	eventType := str(args, "type")
	correlationID := strings.TrimSpace(str(args, "correlationId"))
	idempotencyKey := strings.TrimSpace(str(args, "idempotencyKey"))
	if deps.GatewayMode == appservices.MCPGatewayModeSecure {
		if correlationID == "" {
			return secureMutationFailure("state.correlation_required", "correlationId is required for secure event proposals")
		}
		if idempotencyKey == "" {
			return secureMutationFailure("state.idempotency_required", "idempotencyKey is required for secure event proposals")
		}
	}
	payload := map[string]any{}
	if raw, ok := args["payload"]; ok {
		if m, ok := raw.(map[string]any); ok {
			payload = m
		}
	}
	var (
		evt    *entities.Event
		reused bool
		err    error
	)
	if idempotencyKey != "" {
		idempotent, ok := deps.Orchestrator.(IdempotentOrchestratorPort)
		if !ok {
			if deps.GatewayMode == appservices.MCPGatewayModeSecure {
				return secureMutationFailure("state.idempotency_unavailable", "secure event idempotency is not configured")
			}
		} else {
			evt, reused, err = idempotent.ProposeEventWithIdempotency(ctx, tenantID, instanceID, eventType, payload, correlationID, idempotencyKey)
		}
	}
	if evt == nil && err == nil {
		evt, err = deps.Orchestrator.ProposeEvent(ctx, tenantID, instanceID, eventType, payload, correlationID)
	}
	if err != nil {
		recordRuntimeTraceError(ctx, deps, tenantID, instanceID, entities.RuntimeTraceStageEventHandling, err)
		return toolError(err)
	}
	eventID := evt.EventID
	if eventID == "" {
		eventID = evt.ID
	}
	recordRuntimeTrace(ctx, deps, tenantID, instanceID, appservices.TraceRecordInput{
		Stage:  entities.RuntimeTraceStageEventHandling,
		Status: entities.RuntimeTraceStatusSucceeded,
		Attributes: map[string]any{
			"event_id":   eventID,
			"event_type": eventType,
		},
	})
	return mcp.NewToolResultJSON(map[string]any{
		"ok":        true,
		"eventId":   eventID,
		"eventType": evt.Type,
		"sequence":  evt.Sequence,
		"reused":    reused,
	})
}

func handleStartWorkflow(ctx context.Context, deps Dependencies, tenantID, workflowID, versionID, correlation string) (*mcp.CallToolResult, error) {
	return handleStartWorkflowWithIdempotency(ctx, deps, tenantID, workflowID, versionID, correlation, "")
}

func handleStartWorkflowRequest(ctx context.Context, deps Dependencies, tenantID string, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := getArgs(req)
	return handleStartWorkflowWithIdempotency(ctx, deps, tenantID, str(args, "workflow"), str(args, "version"), str(args, "correlation"), strings.TrimSpace(str(args, "idempotencyKey")))
}

func handleStartWorkflowWithIdempotency(ctx context.Context, deps Dependencies, tenantID, workflowID, versionID, correlation, idempotencyKey string) (*mcp.CallToolResult, error) {
	if deps.Orchestrator == nil {
		return toolUnavailable("orchestrator not configured")
	}
	if deps.GatewayMode == appservices.MCPGatewayModeSecure && idempotencyKey == "" {
		return secureMutationFailure("state.idempotency_required", "idempotencyKey is required for secure workflow starts")
	}
	var (
		inst   *entities.WorkflowInstance
		reused bool
		err    error
	)
	if idempotencyKey != "" {
		idempotent, ok := deps.Orchestrator.(IdempotentOrchestratorPort)
		if !ok {
			if deps.GatewayMode == appservices.MCPGatewayModeSecure {
				return secureMutationFailure("state.idempotency_unavailable", "secure workflow start idempotency is not configured")
			}
		} else {
			inst, reused, err = idempotent.StartWorkflowWithIdempotency(ctx, tenantID, workflowID, versionID, correlation, idempotencyKey)
		}
	}
	if inst == nil && err == nil {
		inst, err = deps.Orchestrator.StartWorkflow(ctx, tenantID, workflowID, versionID, correlation)
	}
	if err != nil {
		return toolError(err)
	}
	recordRuntimeTrace(ctx, deps, tenantID, inst.ID, appservices.TraceRecordInput{
		Stage:         entities.RuntimeTraceStageWorkflowLookup,
		Status:        entities.RuntimeTraceStatusSucceeded,
		CorrelationID: traceStringPtr(correlation),
		Attributes: map[string]any{
			"workflow_id": workflowID,
			"version_id":  versionID,
		},
	})
	return mcp.NewToolResultJSON(map[string]any{
		"ok":         true,
		"instanceId": inst.ID,
		"status":     inst.Status,
		"reused":     reused,
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
			"id":       i.ID,
			"workflow": i.WorkflowID,
			"status":   i.Status,
			"version":  i.Version,
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
			"id":        e.ID,
			"type":      e.Type,
			"sequence":  e.Sequence,
			"source":    e.Source,
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
		"replayed": true,
		"context":  contextSnap,
		"stateKey": stateKey,
	}
	return mcp.NewToolResultJSON(out)
}

func handleCompiledContext(ctx context.Context, deps Dependencies, tenantID string, args map[string]any) (*mcp.CallToolResult, error) {
	if deps.ContextCompiler == nil {
		return toolUnavailable("context compiler not configured")
	}
	compiled, err := deps.ContextCompiler.Compile(ctx, appservices.CompileArgs{
		TenantID:           tenantID,
		ConversationID:     str(args, "conversation"),
		WorkflowInstanceID: str(args, "instance"),
		OwnerType:          str(args, "ownerType"),
		OwnerID:            str(args, "ownerId"),
		Query:              str(args, "query"),
	})
	if err != nil {
		recordRuntimeTraceError(ctx, deps, tenantID, str(args, "instance"), entities.RuntimeTraceStageContextResolution, err)
		return toolError(err)
	}
	if instanceID := str(args, "instance"); instanceID != "" {
		recordRuntimeTrace(ctx, deps, tenantID, instanceID, appservices.TraceRecordInput{
			Stage:         entities.RuntimeTraceStageContextResolution,
			Status:        entities.RuntimeTraceStatusSucceeded,
			CorrelationID: traceStringPtr(str(args, "conversation")),
			Attributes: map[string]any{
				"available_count": len(compiled.Available),
				"missing_count":   len(compiled.Missing),
				"redacted":        compiled.Redacted,
			},
		})
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

func recordRuntimeTrace(ctx context.Context, deps Dependencies, tenantID, instanceID string, input appservices.TraceRecordInput) {
	if deps.TraceRecorder == nil || tenantID == "" || instanceID == "" {
		return
	}
	_, _ = deps.TraceRecorder.Record(ctx, tenantID, instanceID, input)
}

func recordRuntimeTraceError(ctx context.Context, deps Dependencies, tenantID, instanceID string, stage entities.RuntimeTraceStage, err error) {
	if err == nil {
		return
	}
	errorCode := "RUNTIME_STAGE_FAILED"
	reasonCode := "STAGE_FAILED"
	summary := err.Error()
	var guardErr *engine.ErrGuardFailed
	if errors.As(err, &guardErr) {
		errorCode = "GUARD_FAILED"
		reasonCode = "GUARD_FAILED"
	}
	recordRuntimeTrace(ctx, deps, tenantID, instanceID, appservices.TraceRecordInput{
		Stage:      stage,
		Status:     entities.RuntimeTraceStatusFailed,
		ErrorCode:  &errorCode,
		ReasonCode: &reasonCode,
		Summary:    &summary,
	})
}

func stateInstanceID(state *entities.StateInstance) string {
	if state == nil {
		return ""
	}
	return state.ID
}

func traceStringPtr(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

// ---- shared helpers ----

func toolUnavailable(msg string) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultJSON(map[string]any{"ok": false, "message": msg})
}

func toolError(err error) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultJSON(map[string]any{"ok": false, "message": err.Error()})
}
