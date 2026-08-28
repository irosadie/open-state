package repositories

import (
	"context"
	"encoding/json"
	"time"

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
	// ListFiltered returns audit entries for a tenant with optional filters and
	// pagination, newest first (PRD 50).
	ListFiltered(ctx context.Context, tenantID string, filter AuditFilter) ([]entities.AuditLog, error)
	// CountFiltered returns the number of audit entries matching the filters for
	// a tenant, used for pagination (PRD 50).
	CountFiltered(ctx context.Context, tenantID string, filter AuditFilter) (int64, error)
}

// AuditFilter is a tenant-scoped, optional-filter set for listing audit entries
// (PRD 50). Nil/zero filter fields mean "no filter".
type AuditFilter struct {
	Action        *entities.AuditAction
	ResourceType  *string
	ResourceID    *string
	Actor         *string
	CorrelationID *string
	From          *time.Time
	To            *time.Time
	Offset        int
	Limit         int
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
