package services

import (
	"context"
	"encoding/json"
	"log/slog"

	"go.opentelemetry.io/otel/trace"

	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
)

// AuditMetrics is an optional sink for audit-volume metrics (PRD §84).
type AuditMetrics interface {
	IncAudit(action string)
}

// AuditWriter is an application service that records important operations into
// the append-only, tenant-isolated audit trail (PRD 50). Writes are best-effort
// by default: an audit failure SHALL NOT fail the originating business operation
// unless the operation is itself audit-critical.
type AuditWriter struct {
	repo    repositories.IAuditRepository
	logger  *slog.Logger
	metrics AuditMetrics
}

// NewAuditWriter builds an AuditWriter over the audit repository.
func NewAuditWriter(repo repositories.IAuditRepository, logger *slog.Logger, metrics AuditMetrics) *AuditWriter {
	if logger == nil {
		logger = slog.Default()
	}
	return &AuditWriter{repo: repo, logger: logger, metrics: metrics}
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
	// Correlation: prefer the caller-provided correlation id, else the current
	// OTel trace id so audit entries link to distributed traces (PRD §84).
	if correlationID == nil {
		if sc := trace.SpanContextFromContext(ctx); sc.IsValid() && sc.IsSampled() {
			id := sc.TraceID().String()
			correlationID = &id
		}
	}
	if correlationID != nil {
		entry.CorrelationID = correlationID
	}
	if _, err := w.repo.Append(ctx, tenantID, entry); err != nil {
		// Best-effort: do not fail the originating operation on audit failure.
		w.logger.Warn("audit write failed",
			"tenant", tenantID, "action", action, "resource", resourceID, "error", err.Error(),
		)
		return
	}
	if w.metrics != nil {
		w.metrics.IncAudit(string(action))
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
