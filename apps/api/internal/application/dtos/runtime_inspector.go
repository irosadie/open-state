package dtos

// RuntimeWorkflowDTO is the safe workflow/version summary shown to operators.
type RuntimeWorkflowDTO struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	VersionID string `json:"versionId"`
	Version   int    `json:"version"`
}

// RuntimeStateDTO identifies the current or historical state occurrence.
type RuntimeStateDTO struct {
	ID        string  `json:"id"`
	Key       string  `json:"key"`
	Name      string  `json:"name"`
	Status    string  `json:"status"`
	EnteredAt string  `json:"enteredAt"`
	ExitedAt  *string `json:"exitedAt,omitempty"`
}

// RuntimeInstanceSummaryDTO is one tenant-scoped row in instance discovery.
type RuntimeInstanceSummaryDTO struct {
	ID             string             `json:"id"`
	Workflow       RuntimeWorkflowDTO `json:"workflow"`
	Status         string             `json:"status"`
	CurrentState   *RuntimeStateDTO   `json:"currentState,omitempty"`
	CorrelationID  *string            `json:"correlationId,omitempty"`
	LastActivityAt string             `json:"lastActivityAt"`
}

// RuntimeInstanceListDTO is the paginated discovery response.
type RuntimeInstanceListDTO struct {
	Data     []RuntimeInstanceSummaryDTO `json:"data"`
	Page     int                         `json:"page"`
	PageSize int                         `json:"pageSize"`
	Total    int                         `json:"total"`
	HasNext  bool                        `json:"hasNext"`
}

// RuntimeContextDTO contains only sanitized context values and explicit missing
// keys derived from the pinned current state definition.
type RuntimeContextDTO struct {
	Available map[string]any `json:"available"`
	Missing   []string       `json:"missing"`
	Redacted  bool           `json:"redacted"`
}

// RuntimeTimelineEntryDTO is a safe state/event/decision activity entry.
type RuntimeTimelineEntryDTO struct {
	ID            string  `json:"id"`
	Kind          string  `json:"kind"`
	Type          string  `json:"type"`
	Label         string  `json:"label"`
	Status        string  `json:"status"`
	Sequence      int64   `json:"sequence"`
	OccurredAt    string  `json:"occurredAt"`
	CorrelationID *string `json:"correlationId,omitempty"`
	ReasonCode    *string `json:"reasonCode,omitempty"`
}

// RuntimeInstanceDetailDTO is the complete runtime read model for one instance.
type RuntimeInstanceDetailDTO struct {
	Summary             RuntimeInstanceSummaryDTO `json:"summary"`
	CurrentState        *RuntimeStateDTO          `json:"currentState,omitempty"`
	Context             RuntimeContextDTO         `json:"context"`
	Timeline            []RuntimeTimelineEntryDTO `json:"timeline"`
	AuditCorrelationIDs []string                  `json:"auditCorrelationIds"`
}

// RuntimeTraceEntryDTO is the redacted, product-level Debug View entry.
type RuntimeTraceEntryDTO struct {
	ID                string         `json:"id"`
	TurnID            *string        `json:"turnId,omitempty"`
	Sequence          int64          `json:"sequence"`
	Stage             string         `json:"stage"`
	Source            string         `json:"source"`
	Status            string         `json:"status"`
	OccurredAt        string         `json:"occurredAt"`
	CorrelationID     *string        `json:"correlationId,omitempty"`
	DurationMS        *int64         `json:"durationMs,omitempty"`
	ReasonCode        *string        `json:"reasonCode,omitempty"`
	ErrorCode         *string        `json:"errorCode,omitempty"`
	ProviderAlias     *string        `json:"providerAlias,omitempty"`
	ProviderReference *string        `json:"providerReference,omitempty"`
	Summary           *string        `json:"summary,omitempty"`
	Attributes        map[string]any `json:"attributes"`
}

// RuntimeTraceDTO distinguishes absent/unrecorded trace data from an empty
// successful provider result.
type RuntimeTraceDTO struct {
	Available bool                   `json:"available"`
	Data      []RuntimeTraceEntryDTO `json:"data"`
}
