package repositories

import (
	"context"
	"encoding/json"

	"github.com/irosadie/open-state/api/internal/domain/entities"
)

// IAuditRepository defines the persistence contract for the append-only audit
// trail (PRD 50). It is tenant-scoped: every method takes an explicit tenantID
// (PRD 4, 96) so cross-tenant access is impossible at the data-access layer.
// It operates on domain entities (DB-agnostic, ADR-001).
type IAuditRepository interface {
	// Append persists a new audit entry (PRD 50). Entries are append-only.
	Append(ctx context.Context, tenantID string, input AppendAuditLogInput) (*entities.AuditLog, error)
	// ListByTenant returns all audit entries for a tenant, newest first.
	ListByTenant(ctx context.Context, tenantID string) ([]entities.AuditLog, error)
	// ListByAction returns audit entries for a tenant filtered by action (PRD 50).
	ListByAction(ctx context.Context, tenantID string, action entities.AuditAction) ([]entities.AuditLog, error)
	// ListByResource returns audit entries for a tenant filtered by resource.
	ListByResource(ctx context.Context, tenantID, resourceType, resourceID string) ([]entities.AuditLog, error)
}

// AppendAuditLogInput carries the fields needed to append an audit entry.
type AppendAuditLogInput struct {
	Actor         string
	Action        entities.AuditAction
	ResourceType  string
	ResourceID    string
	Before        *json.RawMessage // optional pre-operation state
	After         *json.RawMessage // optional post-operation state
	CorrelationID *string          // optional conversation/business correlation
}
