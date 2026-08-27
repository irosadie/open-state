package entities

import (
	"encoding/json"
	"time"
)

// PolicyScopeType is the scope at which a policy applies (PRD §3.13, §12).
type PolicyScopeType string

const (
	PolicyScopeTenant   PolicyScopeType = "TENANT"
	PolicyScopeWorkflow PolicyScopeType = "WORKFLOW"
	PolicyScopeState    PolicyScopeType = "STATE"
)

// Policy holds runtime/security/business constraints scoped to a
// tenant/workflow/state (PRD §3.13, §12). Content is a JSONB policy document
// (timeout, retry, human_handoff, max_retries, interruptible, etc.).
type Policy struct {
	ID        string
	TenantID  string
	ScopeType PolicyScopeType
	ScopeID   string // tenant/workflow/state id
	Type      string // e.g. timeout / retry / human_handoff / workflow
	Content   json.RawMessage
	CreatedAt time.Time
	UpdatedAt time.Time
}
