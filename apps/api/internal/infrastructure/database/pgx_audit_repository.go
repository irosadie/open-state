package database

import (
	"context"
	"encoding/json"

	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	"github.com/irosadie/open-state/api/internal/infrastructure/db"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/sqlc-dev/pqtype"
)

// PgxAuditRepository implements IAuditRepository on PostgreSQL via sqlc.
type PgxAuditRepository struct {
	queries *db.Queries
}

// NewPgxAuditRepository returns a PostgreSQL-backed IAuditRepository.
func NewPgxAuditRepository(pool *pgxpool.Pool) repositories.IAuditRepository {
	sqlDB := stdlib.OpenDBFromPool(pool)
	return newPgxAuditRepository(db.New(sqlDB))
}

// newPgxAuditRepository builds an IAuditRepository from a sqlc queries handle.
// It enables the composed PostgresAdapter to bind this repository to a shared
// transaction via WithTx.
func newPgxAuditRepository(q *db.Queries) repositories.IAuditRepository {
	return &PgxAuditRepository{queries: q}
}

func (r *PgxAuditRepository) Append(ctx context.Context, tenantID string, input repositories.AppendAuditLogInput) (*entities.AuditLog, error) {
	row, err := r.queries.AppendAuditLog(ctx, db.AppendAuditLogParams{
		TenantID:      mustUUID(tenantID),
		Actor:         input.Actor,
		Action:        string(input.Action),
		ResourceType:  input.ResourceType,
		ResourceID:    input.ResourceID,
		Before:        nullRawMessagePtr(input.Before),
		After:         nullRawMessagePtr(input.After),
		CorrelationID: nullString(input.CorrelationID),
	})
	if err != nil {
		return nil, mapPgError(err, "append audit log")
	}
	return mapAuditLog(row), nil
}

func (r *PgxAuditRepository) ListByTenant(ctx context.Context, tenantID string) ([]entities.AuditLog, error) {
	rows, err := r.queries.ListAuditByTenant(ctx, mustUUID(tenantID))
	if err != nil {
		return nil, err
	}
	out := make([]entities.AuditLog, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapAuditLog(row))
	}
	return out, nil
}

func (r *PgxAuditRepository) ListByAction(ctx context.Context, tenantID string, action entities.AuditAction) ([]entities.AuditLog, error) {
	rows, err := r.queries.ListAuditByAction(ctx, db.ListAuditByActionParams{
		TenantID: mustUUID(tenantID),
		Action:   string(action),
	})
	if err != nil {
		return nil, err
	}
	out := make([]entities.AuditLog, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapAuditLog(row))
	}
	return out, nil
}

func (r *PgxAuditRepository) ListByResource(ctx context.Context, tenantID, resourceType, resourceID string) ([]entities.AuditLog, error) {
	rows, err := r.queries.ListAuditByResource(ctx, db.ListAuditByResourceParams{
		TenantID:     mustUUID(tenantID),
		ResourceType: resourceType,
		ResourceID:   resourceID,
	})
	if err != nil {
		return nil, err
	}
	out := make([]entities.AuditLog, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapAuditLog(row))
	}
	return out, nil
}

// ---- mappers ----

func mapAuditLog(row db.AuditLog) *entities.AuditLog {
	return &entities.AuditLog{
		ID:            row.ID.String(),
		TenantID:      row.TenantID.String(),
		Actor:         row.Actor,
		Action:        entities.AuditAction(row.Action),
		ResourceType:  row.ResourceType,
		ResourceID:    row.ResourceID,
		Before:        nullRawMessageToPtr(row.Before),
		After:         nullRawMessageToPtr(row.After),
		CorrelationID: nullStringPtr(row.CorrelationID),
		OccurredAt:    row.OccurredAt,
		CreatedAt:     row.CreatedAt,
	}
}

// ---- helpers ----

func nullRawMessagePtr(s *json.RawMessage) pqtype.NullRawMessage {
	if s == nil {
		return pqtype.NullRawMessage{Valid: false}
	}
	return pqtype.NullRawMessage{RawMessage: *s, Valid: true}
}

func nullRawMessageToPtr(n pqtype.NullRawMessage) *json.RawMessage {
	if !n.Valid {
		return nil
	}
	raw := json.RawMessage(n.RawMessage)
	return &raw
}
