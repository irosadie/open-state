package entities

import (
	"encoding/json"
	"time"
)

// ContextScopeType identifies the entity a runtime context value is scoped to (PRD §23).
type ContextScopeType string

const (
	ContextScopeTenant            ContextScopeType = "TENANT"
	ContextScopeConversation      ContextScopeType = "CONVERSATION"
	ContextScopeWorkflowInstance  ContextScopeType = "WORKFLOW_INSTANCE"
	ContextScopeStateInstance     ContextScopeType = "STATE_INSTANCE"
)

// ContextRecord is a scoped, typed, versioned runtime context value (PRD §23, §36, §131).
// Values are JSONB (PRD §131) and version enables optimistic updates (PRD §31).
type ContextRecord struct {
	ID        string
	TenantID  string
	ScopeType ContextScopeType
	ScopeID   string
	Key       string          // e.g. booking.time_start
	Value     json.RawMessage // typed value / snapshot (PRD §131)
	Version   int             // optimistic lock (PRD §31)
	CreatedAt time.Time
	UpdatedAt time.Time
}
