package entities

import "time"

// BindingScopeType is the scope at which a capability binding applies (PRD §60).
type BindingScopeType string

const (
	BindingScopeTenant    BindingScopeType = "TENANT"
	BindingScopeWorkflow  BindingScopeType = "WORKFLOW"
	BindingScopeState     BindingScopeType = "STATE"
)

// BindingPermission is the effective permission of a capability binding (PRD §60).
// Resolution uses most-restrictive-wins across overlapping scopes.
type BindingPermission string

const (
	BindingPermissionAllow BindingPermission = "ALLOW"
	BindingPermissionDeny  BindingPermission = "DENY"
)

// CapabilityBinding scopes the availability of a capability to a
// tenant/workflow/state level with most-restrictive-wins resolution (PRD §60).
type CapabilityBinding struct {
	ID           string
	TenantID     string
	CapabilityID string
	ScopeType    BindingScopeType
	ScopeID      string // tenant/workflow/state id
	Permission   BindingPermission
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
