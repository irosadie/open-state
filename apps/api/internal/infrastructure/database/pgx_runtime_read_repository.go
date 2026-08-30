package database

import (
	"context"

	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	"github.com/irosadie/open-state/api/internal/infrastructure/db"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

// PgxRuntimeReadRepository provides the joined definition/state projections
// needed by Runtime Inspector. It exposes domain entities only.
type PgxRuntimeReadRepository struct {
	queries *db.Queries
}

func NewPgxRuntimeReadRepository(pool *pgxpool.Pool) repositories.IRuntimeReadRepository {
	sqlDB := stdlib.OpenDBFromPool(pool)
	return newPgxRuntimeReadRepository(db.New(sqlDB))
}

func newPgxRuntimeReadRepository(q *db.Queries) repositories.IRuntimeReadRepository {
	return &PgxRuntimeReadRepository{queries: q}
}

func (r *PgxRuntimeReadRepository) FindWorkflow(ctx context.Context, tenantID, workflowID string) (*entities.Workflow, error) {
	row, err := r.queries.FindRuntimeWorkflow(ctx, db.FindRuntimeWorkflowParams{
		ID:       mustUUID(workflowID),
		TenantID: mustUUID(tenantID),
	})
	if err != nil {
		return nil, mapNotFound(err, "workflow")
	}
	return mapWorkflow(row), nil
}

func (r *PgxRuntimeReadRepository) FindWorkflowVersion(ctx context.Context, tenantID, workflowVersionID string) (*entities.WorkflowVersion, error) {
	row, err := r.queries.FindRuntimeWorkflowVersion(ctx, db.FindRuntimeWorkflowVersionParams{
		ID:       mustUUID(workflowVersionID),
		TenantID: mustUUID(tenantID),
	})
	if err != nil {
		return nil, mapNotFound(err, "workflow version")
	}
	return mapWorkflowVersion(row), nil
}

func (r *PgxRuntimeReadRepository) ListStatesByVersion(ctx context.Context, tenantID, workflowVersionID string) ([]entities.State, error) {
	rows, err := r.queries.ListRuntimeStatesByVersion(ctx, db.ListRuntimeStatesByVersionParams{
		WorkflowVersionID: mustUUID(workflowVersionID),
		TenantID:          mustUUID(tenantID),
	})
	if err != nil {
		return nil, err
	}
	out := make([]entities.State, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapState(row))
	}
	return out, nil
}

func (r *PgxRuntimeReadRepository) ListStateInstancesByWorkflowInstance(ctx context.Context, tenantID, workflowInstanceID string) ([]entities.StateInstance, error) {
	rows, err := r.queries.ListStateInstancesByWorkflowInstance(ctx, db.ListStateInstancesByWorkflowInstanceParams{
		WorkflowInstanceID: mustUUID(workflowInstanceID),
		TenantID:           mustUUID(tenantID),
	})
	if err != nil {
		return nil, err
	}
	out := make([]entities.StateInstance, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapStateInstance(row))
	}
	return out, nil
}

var _ repositories.IRuntimeReadRepository = (*PgxRuntimeReadRepository)(nil)
