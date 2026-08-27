// Package engine implements the deterministic runtime engine core for OpenState.
//
// This package is domain-pure: it has no HTTP, database, or LLM dependency.
// It speaks to repository ports (interfaces) that a concrete implementation
// (e.g. PostgreSQL) provides later (Epic #3). It executes workflow state
// machines deterministically so the LLM never owns workflow state (PRD §2).
package engine

import "time"

// WorkflowNodeKind identifies the logical role of a node in a workflow graph.
type WorkflowNodeKind string

const (
	NodeKindStart     WorkflowNodeKind = "START"
	NodeKindState     WorkflowNodeKind = "STATE"
	NodeKindDecision  WorkflowNodeKind = "DECISION"
	NodeKindEvent     WorkflowNodeKind = "EVENT"
	NodeKindEnd       WorkflowNodeKind = "END"
)

// WorkflowStatus is the lifecycle of a workflow definition (PRD §9).
type WorkflowStatus string

const (
	WorkflowDraft      WorkflowStatus = "DRAFT"
	WorkflowValidating WorkflowStatus = "VALIDATING"
	WorkflowValid      WorkflowStatus = "VALID"
	WorkflowPublished  WorkflowStatus = "PUBLISHED"
	WorkflowArchived   WorkflowStatus = "ARCHIVED"
)

// GuardOperator is a supported comparison operator (PRD §35).
type GuardOperator string

const (
	OpEq        GuardOperator = "=="
	OpNeq       GuardOperator = "!="
	OpGt        GuardOperator = ">"
	OpGte       GuardOperator = ">="
	OpLt        GuardOperator = "<"
	OpLte       GuardOperator = "<="
	OpIn        GuardOperator = "IN"
	OpNotIn     GuardOperator = "NOT_IN"
	OpExists    GuardOperator = "EXISTS"
	OpNotExists GuardOperator = "NOT_EXISTS"
)

// GuardCondition is a single field/operator/value predicate.
type GuardCondition struct {
	Field    string        `json:"field"`
	Operator GuardOperator `json:"operator"`
	Value    any           `json:"value,omitempty"`
}

// GuardGroup combines conditions with AND/OR logic (PRD §35).
type GuardGroup struct {
	Logic      string           `json:"logic"` // "AND" | "OR"
	Conditions []GuardCondition `json:"conditions"`
}

// TransitionDefinition is a valid movement from one state to another
// based on an event and optional guards (PRD §33).
type TransitionDefinition struct {
	ID             string       `json:"id"`
	SourceStateID  string       `json:"sourceStateId"`
	Event          string       `json:"event"`
	TargetStateID  string       `json:"targetStateId"`
	Guards         []GuardGroup `json:"guards,omitempty"`
	Priority       int          `json:"priority"`
}

// WorkflowNode is a single node (state/decision/start/end/event) in a workflow.
type WorkflowNode struct {
	ID               string            `json:"id"`
	Kind             WorkflowNodeKind  `json:"kind"`
	Name             string            `json:"name"`
	Description      string            `json:"description,omitempty"`
	RequiredContext  []string          `json:"requiredContext,omitempty"`
	Capabilities     []string          `json:"capabilities,omitempty"`
	Instructions     string            `json:"instructions,omitempty"`
	Policy           StatePolicy       `json:"policy,omitempty"`
	IsTerminal       bool              `json:"isTerminal,omitempty"`
}

// StatePolicy holds per-state runtime constraints (PRD §25, §48, §49).
type StatePolicy struct {
	TimeoutSeconds  *int              `json:"timeoutSeconds,omitempty"`
	OnTimeout       string            `json:"onTimeout,omitempty"`
	Retry           *RetryPolicy      `json:"retry,omitempty"`
	HumanHandoff    *HumanHandoff     `json:"humanHandoff,omitempty"`
}

// RetryPolicy configures retry behavior for a state (PRD §48).
type RetryPolicy struct {
	MaxAttempts     int      `json:"maxAttempts"`
	BackoffMs       int      `json:"backoffMs"`
	RetryableEvents []string `json:"retryableEvents,omitempty"`
}

// HumanHandoff marks a state that escalates to a human agent (PRD §49).
type HumanHandoff struct {
	Enabled bool `json:"enabled"`
}

// WorkflowPolicy holds workflow-level runtime constraints (PRD §26, §41, §42).
type WorkflowPolicy struct {
	MaxDurationSeconds *int            `json:"maxDurationSeconds,omitempty"`
	Interruptible      string          `json:"interruptible"` // NEVER|USER_REQUESTED|HIGH_PRIORITY|ALWAYS
	Priority           int             `json:"priority"`
}

// WorkflowTrigger defines how a workflow can be started (PRD §40).
type WorkflowTrigger struct {
	Event  string `json:"event"`
	Source string `json:"source"` // event|api|intent|webhook|schedule
}

// WorkflowDefinition is the full declarative definition of a workflow (PRD §161).
type WorkflowDefinition struct {
	Slug            string               `json:"slug"`
	Name            string               `json:"name"`
	Description     string               `json:"description,omitempty"`
	SchemaVersion   int                  `json:"schemaVersion"`
	Status          WorkflowStatus       `json:"status"`
	EntryNodeID     string               `json:"entryNodeId,omitempty"`
	Nodes           []WorkflowNode       `json:"nodes"`
	Transitions     []TransitionDefinition `json:"transitions"`
	Policy          WorkflowPolicy       `json:"policy"`
	Triggers        []WorkflowTrigger    `json:"triggers,omitempty"`
}

// WorkflowInstanceStatus is the runtime lifecycle of an instance (PRD §10).
type WorkflowInstanceStatus string

const (
	InstanceCreated   WorkflowInstanceStatus = "CREATED"
	InstanceRunning   WorkflowInstanceStatus = "RUNNING"
	InstanceWaiting   WorkflowInstanceStatus = "WAITING"
	InstanceSuspended WorkflowInstanceStatus = "SUSPENDED"
	InstanceCompleted WorkflowInstanceStatus = "COMPLETED"
	InstanceCancelled WorkflowInstanceStatus = "CANCELLED"
	InstanceFailed    WorkflowInstanceStatus = "FAILED"
	InstanceExpired   WorkflowInstanceStatus = "EXPIRED"
)

// WorkflowInstance is a runtime execution of a workflow (PRD §3.4).
type WorkflowInstance struct {
	ID              string                `json:"id"`
	TenantID        string                `json:"tenantId"`
	WorkflowID      string                `json:"workflowId"`
	WorkflowVersionID string              `json:"workflowVersionId"`
	ConversationID  string                `json:"conversationId,omitempty"`
	Status          WorkflowInstanceStatus `json:"status"`
	CurrentStateID  string                `json:"currentStateId,omitempty"`
	Context         map[string]any        `json:"context,omitempty"`
	Version         int                   `json:"version"`
	CreatedAt       time.Time             `json:"createdAt"`
	UpdatedAt       time.Time             `json:"updatedAt"`
}

// StateInstanceStatus is the runtime lifecycle of a state (PRD §11).
type StateInstanceStatus string

const (
	StateEntering   StateInstanceStatus = "ENTERING"
	StateActive     StateInstanceStatus = "ACTIVE"
	StateWaiting    StateInstanceStatus = "WAITING"
	StateExiting    StateInstanceStatus = "EXITING"
	StateCompleted  StateInstanceStatus = "COMPLETED"
	StateFailed     StateInstanceStatus = "FAILED"
	StateExpired    StateInstanceStatus = "EXPIRED"
	StateCancelled  StateInstanceStatus = "CANCELLED"
)

// StateInstance is a runtime occurrence of a state within a workflow instance.
type StateInstance struct {
	ID                string            `json:"id"`
	WorkflowInstanceID string           `json:"workflowInstanceId"`
	StateID           string            `json:"stateId"`
	Status            StateInstanceStatus `json:"status"`
	EnteredAt         time.Time         `json:"enteredAt"`
	ExpiresAt         *time.Time        `json:"expiresAt,omitempty"`
	ExitedAt          *time.Time        `json:"exitedAt,omitempty"`
}

// EventSource identifies where an event originated (PRD §28).
type EventSource string

const (
	SourceUser      EventSource = "USER"
	SourceLLM       EventSource = "LLM"
	SourceMCP       EventSource = "MCP"
	SourceWebhook   EventSource = "WEBHOOK"
	SourceSystem    EventSource = "SYSTEM"
	SourceScheduler EventSource = "SCHEDULER"
	SourceAdmin     EventSource = "ADMIN"
	SourceAPI       EventSource = "API"
)

// Event is something that happened or was requested (PRD §27).
type Event struct {
	ID          string       `json:"id"`
	TenantID    string       `json:"tenantId"`
	Type        string       `json:"type"`
	Source      EventSource  `json:"source"`
	WorkflowInstanceID string `json:"workflowInstanceId,omitempty"`
	Payload     map[string]any `json:"payload,omitempty"`
	Timestamp   time.Time    `json:"timestamp"`
	IdempotencyKey string   `json:"idempotencyKey,omitempty"`
}
