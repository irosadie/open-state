package database

import (
	"context"

	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	"github.com/irosadie/open-state/api/internal/infrastructure/db"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/sqlc-dev/pqtype"
)

type PgxCapabilityEvidenceRepository struct {
	queries *db.Queries
}

func NewPgxCapabilityEvidenceRepository(pool *pgxpool.Pool) repositories.ICapabilityEvidenceRepository {
	return newPgxCapabilityEvidenceRepository(db.New(stdlib.OpenDBFromPool(pool)))
}

func newPgxCapabilityEvidenceRepository(q *db.Queries) repositories.ICapabilityEvidenceRepository {
	return &PgxCapabilityEvidenceRepository{queries: q}
}

func (r *PgxCapabilityEvidenceRepository) Upsert(ctx context.Context, input repositories.CapabilityEvidenceInput) (*entities.CapabilityExecutionEvidence, error) {
	row, err := r.queries.UpsertCapabilityExecutionEvidence(ctx, db.UpsertCapabilityExecutionEvidenceParams{
		TenantID: mustUUID(input.TenantID), ProjectID: mustUUID(input.ProjectID),
		WorkflowInstanceID: mustUUID(input.WorkflowInstanceID), StateID: input.StateID,
		CapabilityID: mustUUID(input.CapabilityID), CapabilityName: input.CapabilityName,
		ProviderServer: input.ProviderServer, ProviderTool: input.ProviderTool,
		CorrelationID: nullString(input.CorrelationID), IdempotencyKey: input.IdempotencyKey,
		Status: string(input.Status), Result: input.Result,
		Error: pqtype.NullRawMessage{RawMessage: input.Error, Valid: len(input.Error) > 0},
	})
	if err != nil {
		return nil, mapPgError(err, "upsert capability evidence")
	}
	return mapCapabilityEvidence(row), nil
}

func (r *PgxCapabilityEvidenceRepository) FindByIdempotency(ctx context.Context, tenantID, projectID, instanceID, stateID, capabilityID, idempotencyKey string) (*entities.CapabilityExecutionEvidence, error) {
	row, err := r.queries.FindCapabilityEvidenceByIdempotency(ctx, db.FindCapabilityEvidenceByIdempotencyParams{
		TenantID: mustUUID(tenantID), ProjectID: mustUUID(projectID), WorkflowInstanceID: mustUUID(instanceID),
		StateID: stateID, CapabilityID: mustUUID(capabilityID), IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		return nil, mapNotFound(err, "capability evidence")
	}
	return mapCapabilityEvidence(row), nil
}

func (r *PgxCapabilityEvidenceRepository) ListByState(ctx context.Context, tenantID, projectID, instanceID, stateID string) ([]entities.CapabilityExecutionEvidence, error) {
	rows, err := r.queries.ListCapabilityEvidenceByState(ctx, db.ListCapabilityEvidenceByStateParams{
		TenantID: mustUUID(tenantID), ProjectID: mustUUID(projectID), WorkflowInstanceID: mustUUID(instanceID), StateID: stateID,
	})
	if err != nil {
		return nil, err
	}
	out := make([]entities.CapabilityExecutionEvidence, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapCapabilityEvidence(row))
	}
	return out, nil
}

func (r *PgxCapabilityEvidenceRepository) ListByInstanceState(ctx context.Context, tenantID, instanceID, stateID string) ([]entities.CapabilityExecutionEvidence, error) {
	rows, err := r.queries.ListCapabilityEvidenceByInstanceState(ctx, db.ListCapabilityEvidenceByInstanceStateParams{
		TenantID: mustUUID(tenantID), WorkflowInstanceID: mustUUID(instanceID), StateID: stateID,
	})
	if err != nil {
		return nil, err
	}
	out := make([]entities.CapabilityExecutionEvidence, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapCapabilityEvidence(row))
	}
	return out, nil
}

func mapCapabilityEvidence(row db.CapabilityExecutionEvidence) *entities.CapabilityExecutionEvidence {
	var correlation *string
	if row.CorrelationID.Valid {
		correlation = &row.CorrelationID.String
	}
	out := &entities.CapabilityExecutionEvidence{
		ID: row.ID.String(), TenantID: row.TenantID.String(), ProjectID: row.ProjectID.String(),
		WorkflowInstanceID: row.WorkflowInstanceID.String(), StateID: row.StateID, CapabilityID: row.CapabilityID.String(),
		CapabilityName: row.CapabilityName, ProviderServer: row.ProviderServer, ProviderTool: row.ProviderTool,
		CorrelationID: correlation, IdempotencyKey: row.IdempotencyKey, Status: entities.CapabilityEvidenceStatus(row.Status),
		Result: row.Result, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
	if row.Error.Valid {
		out.Error = row.Error.RawMessage
	}
	return out
}

var _ repositories.ICapabilityEvidenceRepository = (*PgxCapabilityEvidenceRepository)(nil)
