package dtos

import "encoding/json"

// Workflow Builder API request/response DTOs (PRD 146). All operations are
// tenant+project scoped; tenant comes from the X-Tenant-ID header, never the
// body (PRD §74, §96).

// CreateWorkflowRequest is the payload to create a workflow definition draft.
type CreateWorkflowRequest struct {
	ProjectID   string          `json:"projectId"`
	Slug        string          `json:"slug"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Definition  json.RawMessage `json:"definition"`
}

// UpdateWorkflowRequest is the payload to update a workflow draft's mutable
// fields using optimistic concurrency (PRD §31).
type UpdateWorkflowRequest struct {
	Name        string          `json:"name"`
	Description *string         `json:"description"`
	Definition  json.RawMessage `json:"definition"`
	Version     int             `json:"version"`
}

// PublishWorkflowRequest identifies the current optimistic workflow revision to
// publish; the server snapshots its persisted draft definition.
type PublishWorkflowRequest struct {
	Version int `json:"version"`
}

// WorkflowDTO is the serializable workflow definition root returned to callers.
type WorkflowDTO struct {
	ID             string          `json:"id"`
	TenantID       string          `json:"tenantId"`
	ProjectID      string          `json:"projectId"`
	Slug           string          `json:"slug"`
	Name           string          `json:"name"`
	Description    *string         `json:"description"`
	Status         string          `json:"status"`
	CurrentVersion int             `json:"currentVersion"`
	Version        int             `json:"version"`
	Definition     json.RawMessage `json:"definition"`
	CreatedAt      string          `json:"createdAt"`
	UpdatedAt      string          `json:"updatedAt"`
}

// WorkflowListDTO wraps a tenant/project-scoped workflow list.
type WorkflowListDTO struct {
	Data []WorkflowDTO `json:"data"`
}

// WorkflowVersionDTO is the serializable immutable workflow version.
type WorkflowVersionDTO struct {
	ID         string          `json:"id"`
	WorkflowID string          `json:"workflowId"`
	VersionNo  int             `json:"versionNo"`
	Status     string          `json:"status"`
	IsCurrent  bool            `json:"isCurrent"`
	CreatedAt  string          `json:"createdAt"`
	UpdatedAt  string          `json:"updatedAt"`
	Definition json.RawMessage `json:"definition"`
}

// WorkflowValidationIssue is an authoritative graph validation problem.
type WorkflowValidationIssue struct {
	Code         string `json:"code"`
	Message      string `json:"message"`
	NodeID       string `json:"nodeId,omitempty"`
	TransitionID string `json:"transitionId,omitempty"`
}

// WorkflowDiffItem identifies an item in a semantic graph diff.
type WorkflowDiffItem struct {
	ID            string          `json:"id"`
	Definition    json.RawMessage `json:"definition,omitempty"`
	ChangedFields []string        `json:"changedFields,omitempty"`
}

// WorkflowDiffDTO is a deterministic comparison of two immutable graph snapshots.
type WorkflowDiffDTO struct {
	WorkflowID    string            `json:"workflowId"`
	BaseVersion   int               `json:"baseVersion"`
	TargetVersion int               `json:"targetVersion"`
	Nodes         WorkflowDiffGroup `json:"nodes"`
	Transitions   WorkflowDiffGroup `json:"transitions"`
}

// WorkflowDiffGroup contains added, removed, and changed graph items.
type WorkflowDiffGroup struct {
	Added   []WorkflowDiffItem `json:"added"`
	Removed []WorkflowDiffItem `json:"removed"`
	Changed []WorkflowDiffItem `json:"changed"`
}

// SimulateWorkflowRequest executes an unsaved workflow snapshot in the
// in-memory sandbox. Context is intentionally dynamic JSON because workflow
// guards address tenant-defined fields and nested values.
type SimulateWorkflowRequest struct {
	Definition     json.RawMessage            `json:"definition"`
	InitialContext map[string]json.RawMessage `json:"initialContext"`
	Events         []SimulationEvent          `json:"events"`
}

// SimulationEvent is one operator-supplied event in deterministic order.
type SimulationEvent struct {
	Type    string                     `json:"type"`
	Payload map[string]json.RawMessage `json:"payload"`
}

// SimulationStateDTO identifies the workflow state shown in a trace step.
type SimulationStateDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Kind string `json:"kind"`
}

// SimulationCandidateDTO is the response-safe result of one guard candidate.
type SimulationCandidateDTO struct {
	TransitionID string `json:"transitionId"`
	Event        string `json:"event"`
	Priority     int    `json:"priority"`
	GuardPassed  bool   `json:"guardPassed"`
	GuardError   string `json:"guardError,omitempty"`
}

// SimulationCapabilityRequestDTO describes a capability requested by a state.
// It is always a plan in this endpoint; no provider is invoked.
type SimulationCapabilityRequestDTO struct {
	Name   string `json:"name"`
	Mock   bool   `json:"mock"`
	Status string `json:"status"`
}

// SimulationStepDTO contains one entry or event step from the sandbox trace.
type SimulationStepDTO struct {
	Sequence             int                              `json:"sequence"`
	Outcome              string                           `json:"outcome"`
	EventType            string                           `json:"eventType,omitempty"`
	EventPayload         map[string]json.RawMessage       `json:"eventPayload,omitempty"`
	StateBefore          SimulationStateDTO               `json:"stateBefore"`
	StateAfter           *SimulationStateDTO              `json:"stateAfter,omitempty"`
	Candidates           []SimulationCandidateDTO         `json:"candidates"`
	SelectedTransitionID string                           `json:"selectedTransitionId,omitempty"`
	Context              map[string]json.RawMessage       `json:"context"`
	CapabilityRequests   []SimulationCapabilityRequestDTO `json:"capabilityRequests"`
	ErrorCode            string                           `json:"errorCode,omitempty"`
	ErrorMessage         string                           `json:"errorMessage,omitempty"`
}

// SimulationResultDTO is the complete transient sandbox result.
type SimulationResultDTO struct {
	FinalState   SimulationStateDTO         `json:"finalState"`
	FinalContext map[string]json.RawMessage `json:"finalContext"`
	FinalStatus  string                     `json:"finalStatus"`
	Steps        []SimulationStepDTO        `json:"steps"`
}
