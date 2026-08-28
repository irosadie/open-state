package services

import (
	"context"
	"encoding/json"
	"log"

	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
)

// AuditWriter is an application service that records important operations into
// the append-only, tenant-isolated audit trail (PRD 50). Writes are best-effort
// by default: an audit failure SHALL NOT fail the originating business operation
// unless the operation is itself audit-critical.
type AuditWriter struct {
	repo repositories.IAuditRepository
}

// NewAuditWriter builds an AuditWriter over the audit repository.
func NewAuditWriter(repo repositories.IAuditRepository) *AuditWriter {
	return &AuditWriter{repo: repo}
}

// Write appends a single audit entry for an important operation (PRD 50).
func (w *AuditWriter) Write(ctx context.Context, tenantID, actor string, action entities.AuditAction, resourceType, resourceID string, before, after any, correlationID *string) {
	entry := repositories.AppendAuditLogInput{
		Actor:        actor,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Before:       marshalRaw(before),
		After:        marshalRaw(after),
	}
	if correlationID != nil {
		entry.CorrelationID = correlationID
	}
	if _, err := w.repo.Append(ctx, tenantID, entry); err != nil {
		// Best-effort: do not fail the originating operation on audit failure.
		log.Printf("audit write failed (tenant=%s action=%s resource=%s): %v", tenantID, action, resourceID, err)
	}
}

// marshalRaw marshals an arbitrary value into a JSON RawMessage, returning nil on
// marshalling failure (the raw value is then omitted from the audit entry).
func marshalRaw(v any) *json.RawMessage {
	if v == nil {
		return nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	raw := json.RawMessage(b)
	return &raw
}
