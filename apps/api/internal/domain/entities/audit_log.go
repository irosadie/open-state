package entities

import (
	"encoding/json"
	"time"
)

// AuditAction is the PRD 50 audit event set, recorded on each audit entry. Stored
// as VARCHAR in the DB and constrained to these typed constants at the application
// layer (VARCHAR + Go typed constants, not PostgreSQL ENUM).
type AuditAction string

const (
	AuditActionWorkflowPublished   AuditAction = "workflow.published"
	AuditActionStateEntered        AuditAction = "state.entered"
	AuditActionTransitionExecuted  AuditAction = "transition.executed"
	AuditActionGuardFailed         AuditAction = "guard.failed"
	AuditActionCapabilityInvoked   AuditAction = "capability.invoked"
	AuditActionCapabilityDenied    AuditAction = "capability.denied"
	AuditActionWorkflowSuspended   AuditAction = "workflow.suspended"
	AuditActionWorkflowResumed     AuditAction = "workflow.resumed"
	AuditActionHumanHandoffCreated AuditAction = "human_handoff.created"

	// RBAC actions (PRD 80, 81): role-assignment mutations and authorization denials.
	AuditActionRoleAssigned    AuditAction = "rbac.role_assigned"
	AuditActionRoleUpdated     AuditAction = "rbac.role_updated"
	AuditActionRoleRemoved     AuditAction = "rbac.role_removed"
	AuditActionAuthDenied      AuditAction = "authorization.denied"
	AuditActionTenantUpdated   AuditAction = "tenant.updated"
	AuditActionWorkflowRetried AuditAction = "workflow.retried"

	// Binding actions (PRD 60): capability-binding mutations.
	AuditActionBindingCreated AuditAction = "binding.created"
	AuditActionBindingDeleted AuditAction = "binding.deleted"

	// SSO login (PRD 79).
	AuditActionSSOLogin AuditAction = "auth.sso_login"

	// State MCP machine credential lifecycle and authorization decisions.
	AuditActionAPIKeyCreated AuditAction = "api_key.created"
	AuditActionAPIKeyRevoked AuditAction = "api_key.revoked"
	AuditActionAPIKeyUsed    AuditAction = "api_key.used"
	AuditActionAPIKeyDenied  AuditAction = "api_key.denied"
)

// AuditLog is an append-only, tenant-isolated audit entry (PRD 50). It records an
// actor, action, resource, and the before/after state of an important operation.
// Rows are never updated or deleted during normal operation.
type AuditLog struct {
	ID            string
	TenantID      string
	Actor         string      // user/system id
	Action        AuditAction // PRD 50 audit event set
	ResourceType  string      // workflow / instance / state / event / capability / ...
	ResourceID    string
	Before        *json.RawMessage // state before the operation
	After         *json.RawMessage // state after the operation
	CorrelationID *string          // conversation/business correlation
	OccurredAt    time.Time
	CreatedAt     time.Time
}
