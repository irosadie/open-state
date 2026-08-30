package entities

import (
	"encoding/json"
	"time"
)

// RuntimeTraceStage identifies an observable orchestration boundary. The list is
// intentionally product-oriented; it is not an OpenTelemetry span taxonomy.
type RuntimeTraceStage string

const (
	RuntimeTraceStageIntentResolution    RuntimeTraceStage = "INTENT_RESOLUTION"
	RuntimeTraceStageWorkflowLookup      RuntimeTraceStage = "WORKFLOW_LOOKUP"
	RuntimeTraceStageStateLookup         RuntimeTraceStage = "STATE_LOOKUP"
	RuntimeTraceStageContextResolution   RuntimeTraceStage = "CONTEXT_RESOLUTION"
	RuntimeTraceStageRAGIntegration      RuntimeTraceStage = "RAG_INTEGRATION"
	RuntimeTraceStageMCPActivity         RuntimeTraceStage = "MCP_ACTIVITY"
	RuntimeTraceStageLLMIntegration      RuntimeTraceStage = "LLM_INTEGRATION"
	RuntimeTraceStageEventHandling       RuntimeTraceStage = "EVENT_HANDLING"
	RuntimeTraceStageGuardEvaluation     RuntimeTraceStage = "GUARD_EVALUATION"
	RuntimeTraceStageTransitionSelection RuntimeTraceStage = "TRANSITION_SELECTION"
)

// RuntimeTraceSource identifies whether OpenState or an independent provider
// observed the trace boundary.
type RuntimeTraceSource string

const (
	RuntimeTraceSourceOpenState        RuntimeTraceSource = "OPENSTATE"
	RuntimeTraceSourceExternalProvider RuntimeTraceSource = "EXTERNAL_PROVIDER"
)

// RuntimeTraceStatus is deliberately explicit so an absent entry is never
// interpreted as a successful or failed provider operation.
type RuntimeTraceStatus string

const (
	RuntimeTraceStatusStarted     RuntimeTraceStatus = "STARTED"
	RuntimeTraceStatusSucceeded   RuntimeTraceStatus = "SUCCEEDED"
	RuntimeTraceStatusFailed      RuntimeTraceStatus = "FAILED"
	RuntimeTraceStatusNotRecorded RuntimeTraceStatus = "NOT_RECORDED"
)

// SanitizedAttributes is the only structured payload permitted on a runtime
// trace entry. Values must pass through trace redaction before persistence.
type SanitizedAttributes map[string]any

// RuntimeTraceEntry is an append-only, tenant-scoped product debug record.
type RuntimeTraceEntry struct {
	ID                 string
	TenantID           string
	WorkflowInstanceID string
	TurnID             *string
	Sequence           int64
	Stage              RuntimeTraceStage
	Source             RuntimeTraceSource
	Status             RuntimeTraceStatus
	OccurredAt         time.Time
	CorrelationID      *string
	DurationMS         *int64
	ReasonCode         *string
	ErrorCode          *string
	ProviderAlias      *string
	ProviderReference  *string
	Summary            *string
	Attributes         SanitizedAttributes
}

// MarshalSanitizedAttributes keeps the persistence boundary explicit and makes
// it impossible for callers to pass arbitrary request/response bytes directly.
func (e RuntimeTraceEntry) MarshalSanitizedAttributes() (json.RawMessage, error) {
	if e.Attributes == nil {
		return json.RawMessage(`{}`), nil
	}
	return json.Marshal(e.Attributes)
}
