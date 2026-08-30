package database

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	"github.com/irosadie/open-state/api/internal/infrastructure/db"
	"github.com/irosadie/open-state/go-shared/domain"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

// PgxRuntimeTraceRepository implements the append-only runtime trace port.
type PgxRuntimeTraceRepository struct {
	queries *db.Queries
}

func NewPgxRuntimeTraceRepository(pool *pgxpool.Pool) repositories.IRuntimeTraceRepository {
	sqlDB := stdlib.OpenDBFromPool(pool)
	return newPgxRuntimeTraceRepository(db.New(sqlDB))
}

func newPgxRuntimeTraceRepository(q *db.Queries) repositories.IRuntimeTraceRepository {
	return &PgxRuntimeTraceRepository{queries: q}
}

func (r *PgxRuntimeTraceRepository) Append(ctx context.Context, tenantID string, input repositories.AppendRuntimeTraceInput) (*entities.RuntimeTraceEntry, error) {
	tid, err := parseRuntimeUUID(tenantID, "tenant")
	if err != nil {
		return nil, err
	}
	iid, err := parseRuntimeUUID(input.WorkflowInstanceID, "workflow instance")
	if err != nil {
		return nil, err
	}
	attributes, err := json.Marshal(input.Attributes)
	if err != nil {
		return nil, domain.NewValidation("invalid trace attributes")
	}
	if len(attributes) == 0 || string(attributes) == "null" {
		attributes = []byte("{}")
	}
	row, err := r.queries.AppendRuntimeTraceEntry(ctx, db.AppendRuntimeTraceEntryParams{
		TenantID:           tid,
		WorkflowInstanceID: iid,
		TurnID:             runtimeNullString(input.TurnID),
		Stage:              string(input.Stage),
		Source:             string(input.Source),
		Status:             string(input.Status),
		OccurredAt:         input.OccurredAt,
		CorrelationID:      runtimeNullString(input.CorrelationID),
		DurationMs:         runtimeNullInt64(input.DurationMS),
		ReasonCode:         runtimeNullString(input.ReasonCode),
		ErrorCode:          runtimeNullString(input.ErrorCode),
		ProviderAlias:      runtimeNullString(input.ProviderAlias),
		ProviderReference:  runtimeNullString(input.ProviderReference),
		Summary:            runtimeNullString(input.Summary),
		Attributes:         attributes,
	})
	if err != nil {
		return nil, mapPgError(err, "append runtime trace")
	}
	return mapRuntimeTraceEntry(row), nil
}

func (r *PgxRuntimeTraceRepository) ListByInstance(ctx context.Context, tenantID, workflowInstanceID string) ([]entities.RuntimeTraceEntry, error) {
	rows, err := r.queries.ListRuntimeTraceByInstance(ctx, db.ListRuntimeTraceByInstanceParams{
		TenantID:           mustUUID(tenantID),
		WorkflowInstanceID: mustUUID(workflowInstanceID),
	})
	if err != nil {
		return nil, err
	}
	return mapRuntimeTraceEntries(rows), nil
}

func (r *PgxRuntimeTraceRepository) ListByTurn(ctx context.Context, tenantID, workflowInstanceID, turnID string) ([]entities.RuntimeTraceEntry, error) {
	rows, err := r.queries.ListRuntimeTraceByTurn(ctx, db.ListRuntimeTraceByTurnParams{
		TenantID:           mustUUID(tenantID),
		WorkflowInstanceID: mustUUID(workflowInstanceID),
		TurnID:             runtimeNullString(&turnID),
	})
	if err != nil {
		return nil, err
	}
	return mapRuntimeTraceEntries(rows), nil
}

func mapRuntimeTraceEntries(rows []db.RuntimeTraceEntry) []entities.RuntimeTraceEntry {
	out := make([]entities.RuntimeTraceEntry, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapRuntimeTraceEntry(row))
	}
	return out
}

func mapRuntimeTraceEntry(row db.RuntimeTraceEntry) *entities.RuntimeTraceEntry {
	entry := &entities.RuntimeTraceEntry{
		ID:                 row.ID.String(),
		TenantID:           row.TenantID.String(),
		WorkflowInstanceID: row.WorkflowInstanceID.String(),
		Sequence:           row.Sequence,
		Stage:              entities.RuntimeTraceStage(row.Stage),
		Source:             entities.RuntimeTraceSource(row.Source),
		Status:             entities.RuntimeTraceStatus(row.Status),
		OccurredAt:         row.OccurredAt,
		Attributes:         entities.SanitizedAttributes{},
	}
	entry.TurnID = runtimeStringPtr(row.TurnID)
	entry.CorrelationID = runtimeStringPtr(row.CorrelationID)
	entry.DurationMS = runtimeInt64Ptr(row.DurationMs)
	entry.ReasonCode = runtimeStringPtr(row.ReasonCode)
	entry.ErrorCode = runtimeStringPtr(row.ErrorCode)
	entry.ProviderAlias = runtimeStringPtr(row.ProviderAlias)
	entry.ProviderReference = runtimeStringPtr(row.ProviderReference)
	entry.Summary = runtimeStringPtr(row.Summary)
	if len(row.Attributes) > 0 {
		_ = json.Unmarshal(row.Attributes, &entry.Attributes)
	}
	return entry
}

func parseRuntimeUUID(value, label string) (uuid.UUID, error) {
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, domain.NewValidation("invalid " + label + " id")
	}
	return id, nil
}

func runtimeNullString(value *string) sql.NullString {
	if value == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *value, Valid: true}
}

func runtimeStringPtr(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

func runtimeNullInt64(value *int64) sql.NullInt64 {
	if value == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *value, Valid: true}
}

func runtimeInt64Ptr(value sql.NullInt64) *int64 {
	if !value.Valid {
		return nil
	}
	return &value.Int64
}

var _ repositories.IRuntimeTraceRepository = (*PgxRuntimeTraceRepository)(nil)
