package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	"github.com/irosadie/open-state/api/internal/infrastructure/db"
	domain "github.com/irosadie/open-state/go-shared/domain"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

// PgxInstanceRepository implements IInstanceRepository on PostgreSQL via sqlc.
type PgxInstanceRepository struct {
	queries *db.Queries
	db      *sql.DB
}

// NewPgxInstanceRepository returns a PostgreSQL-backed IInstanceRepository.
func NewPgxInstanceRepository(pool *pgxpool.Pool) repositories.IInstanceRepository {
	sqlDB := stdlib.OpenDBFromPool(pool)
	return newPgxInstanceRepository(db.New(sqlDB), sqlDB)
}

// newPgxInstanceRepository builds an IInstanceRepository from a sqlc queries
// handle and the underlying *sql.DB (used for the repo's own standalone
// transactions). It enables the composed PostgresAdapter to bind this repository
// to a shared transaction via WithTx.
func newPgxInstanceRepository(q *db.Queries, sqlDB *sql.DB) repositories.IInstanceRepository {
	return &PgxInstanceRepository{queries: q, db: sqlDB}
}

func (r *PgxInstanceRepository) Create(ctx context.Context, tenantID string, input repositories.CreateWorkflowInstanceInput) (*entities.WorkflowInstance, error) {
	row, err := r.queries.CreateWorkflowInstance(ctx, db.CreateWorkflowInstanceParams{
		TenantID:          mustUUID(tenantID),
		WorkflowID:        mustUUID(input.WorkflowID),
		WorkflowVersionID: mustUUID(input.WorkflowVersionID),
		CorrelationKey:    nullString(input.CorrelationKey),
		Status:            string(entities.WorkflowInstanceCreated),
		StartedAt:         optTime(input.StartedAt),
		ExpiresAt:         optTime(input.ExpiresAt),
	})
	if err != nil {
		return nil, mapPgError(err, "create workflow instance")
	}
	return mapWorkflowInstance(row), nil
}

func (r *PgxInstanceRepository) FindByID(ctx context.Context, tenantID, id string) (*entities.WorkflowInstance, error) {
	row, err := r.queries.FindWorkflowInstanceByID(ctx, db.FindWorkflowInstanceByIDParams{
		ID:       mustUUID(id),
		TenantID: mustUUID(tenantID),
	})
	if err != nil {
		return nil, mapNotFound(err, "workflow instance")
	}
	return mapWorkflowInstance(row), nil
}

func (r *PgxInstanceRepository) ListByTenant(ctx context.Context, tenantID string) ([]entities.WorkflowInstance, error) {
	rows, err := r.queries.ListWorkflowInstancesByTenant(ctx, mustUUID(tenantID))
	if err != nil {
		return nil, err
	}
	out := make([]entities.WorkflowInstance, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapWorkflowInstance(row))
	}
	return out, nil
}

func (r *PgxInstanceRepository) UpdateStatus(ctx context.Context, tenantID, id string, status entities.WorkflowInstanceStatus, expectedVersion int) (*entities.WorkflowInstance, error) {
	row, err := r.queries.UpdateWorkflowInstanceStatus(ctx, db.UpdateWorkflowInstanceStatusParams{
		ID:       mustUUID(id),
		TenantID: mustUUID(tenantID),
		Status:   string(status),
		Version:  int32(expectedVersion),
	})
	if err != nil {
		return nil, mapOptimisticError(err)
	}
	return mapWorkflowInstance(row), nil
}

// Transition atomically exits the current state instance, inserts the new state
// instance, points the workflow instance's current state at it, and increments the
// parent workflow instance version in one DB transaction (PRD §69).
func (r *PgxInstanceRepository) Transition(ctx context.Context, tenantID string, input repositories.TransitionInput) (*entities.StateInstance, error) {
	tid, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, domain.NewValidation("invalid tenant id")
	}
	wfInstID, err := uuid.Parse(input.WorkflowInstanceID)
	if err != nil {
		return nil, domain.NewValidation("invalid workflow instance id")
	}
	exitID, err := uuid.Parse(input.ExitStateInstanceID)
	if err != nil {
		return nil, domain.NewValidation("invalid exiting state instance id")
	}
	wvID, err := uuid.Parse(input.NewWorkflowVersionID)
	if err != nil {
		return nil, domain.NewValidation("invalid workflow version id")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, domain.NewInternal(err.Error())
	}
	defer tx.Rollback() //nolint:errcheck // rollback is a no-op after commit

	q := r.queries.WithTx(tx)

	// 1. Exit the old (current) state instance with optimistic locking.
	if _, err := q.ExitStateInstance(ctx, db.ExitStateInstanceParams{
		ID:       exitID,
		TenantID: tid,
		Status:   string(entities.StateInstanceExiting),
		Version:  int32(input.ExpectedExitVersion),
	}); err != nil {
		return nil, mapOptimisticError(err)
	}

	// 2. Insert the new state instance (ENTERING).
	newState, err := q.CreateStateInstance(ctx, db.CreateStateInstanceParams{
		TenantID:           tid,
		WorkflowInstanceID: wfInstID,
		WorkflowVersionID:  wvID,
		StateKey:           input.NewStateKey,
		StateID:            optUUID(input.NewStateID),
		Status:             string(entities.StateInstanceEntering),
	})
	if err != nil {
		return nil, mapPgError(err, "create state instance")
	}

	// 3. Point the workflow instance's current state at the new state instance (PRD §7).
	if err := q.SetCurrentStateInstance(ctx, db.SetCurrentStateInstanceParams{
		ID:                     wfInstID,
		TenantID:               tid,
		CurrentStateInstanceID: uuid.NullUUID{UUID: newState.ID, Valid: true},
	}); err != nil {
		return nil, domain.NewInternal(err.Error())
	}

	// 4. Bump the parent workflow instance version (optimistic lock, PRD §69).
	if _, err := q.IncrementWorkflowInstanceVersion(ctx, db.IncrementWorkflowInstanceVersionParams{
		ID:       wfInstID,
		TenantID: tid,
		Version:  int32(input.ExpectedWorkflowVersion),
	}); err != nil {
		return nil, mapOptimisticError(err)
	}

	if err := tx.Commit(); err != nil {
		return nil, domain.NewInternal(err.Error())
	}

	return mapStateInstance(newState), nil
}

func (r *PgxInstanceRepository) CreateStateInstance(ctx context.Context, tenantID string, input repositories.CreateStateInstanceInput) (*entities.StateInstance, error) {
	row, err := r.queries.CreateStateInstance(ctx, db.CreateStateInstanceParams{
		TenantID:           mustUUID(tenantID),
		WorkflowInstanceID: mustUUID(input.WorkflowInstanceID),
		WorkflowVersionID:  mustUUID(input.WorkflowVersionID),
		StateKey:           input.StateKey,
		StateID:            optUUID(input.StateID),
		Status:             string(entities.StateInstanceEntering),
	})
	if err != nil {
		return nil, mapPgError(err, "create state instance")
	}
	return mapStateInstance(row), nil
}

func (r *PgxInstanceRepository) FindStateInstanceByID(ctx context.Context, tenantID, id string) (*entities.StateInstance, error) {
	row, err := r.queries.FindStateInstanceByID(ctx, db.FindStateInstanceByIDParams{
		ID:       mustUUID(id),
		TenantID: mustUUID(tenantID),
	})
	if err != nil {
		return nil, mapNotFound(err, "state instance")
	}
	return mapStateInstance(row), nil
}

func (r *PgxInstanceRepository) UpdateStateInstanceStatus(ctx context.Context, tenantID, id string, status entities.StateInstanceStatus, expectedVersion int) (*entities.StateInstance, error) {
	row, err := r.queries.UpdateStateInstanceStatus(ctx, db.UpdateStateInstanceStatusParams{
		ID:       mustUUID(id),
		TenantID: mustUUID(tenantID),
		Status:   string(status),
		Version:  int32(expectedVersion),
	})
	if err != nil {
		return nil, mapOptimisticError(err)
	}
	return mapStateInstance(row), nil
}

func (r *PgxInstanceRepository) IncrementRetry(ctx context.Context, tenantID, id string, expectedVersion int) (*entities.StateInstance, error) {
	row, err := r.queries.UpdateStateInstanceRetry(ctx, db.UpdateStateInstanceRetryParams{
		ID:       mustUUID(id),
		TenantID: mustUUID(tenantID),
		Version:  int32(expectedVersion),
	})
	if err != nil {
		return nil, mapOptimisticError(err)
	}
	return mapStateInstance(row), nil
}

// ---- mappers ----

func mapWorkflowInstance(row db.WorkflowInstance) *entities.WorkflowInstance {
	return &entities.WorkflowInstance{
		ID:                    row.ID.String(),
		TenantID:              row.TenantID.String(),
		WorkflowID:            row.WorkflowID.String(),
		WorkflowVersionID:     row.WorkflowVersionID.String(),
		CorrelationKey:        row.CorrelationKey,
		Status:                entities.WorkflowInstanceStatus(row.Status),
		Version:               int(row.Version),
		CurrentStateInstanceID: nullUUIDPtr(row.CurrentStateInstanceID),
		StartedAt:             nullTimePtr(row.StartedAt),
		CompletedAt:           nullTimePtr(row.CompletedAt),
		ExpiresAt:             nullTimePtr(row.ExpiresAt),
		CreatedAt:             row.CreatedAt,
		UpdatedAt:             row.UpdatedAt,
	}
}

func mapStateInstance(row db.StateInstance) *entities.StateInstance {
	return &entities.StateInstance{
		ID:                 row.ID.String(),
		TenantID:           row.TenantID.String(),
		WorkflowInstanceID: row.WorkflowInstanceID.String(),
		WorkflowVersionID:  row.WorkflowVersionID.String(),
		StateKey:           row.StateKey,
		StateID:            nullUUIDPtr(row.StateID),
		Status:             entities.StateInstanceStatus(row.Status),
		Version:            int(row.Version),
		RetryCount:         int(row.RetryCount),
		EnteredAt:          row.EnteredAt,
		ExpiresAt:          nullTimePtr(row.ExpiresAt),
		ExitedAt:           nullTimePtr(row.ExitedAt),
		CreatedAt:          row.CreatedAt,
		UpdatedAt:          row.UpdatedAt,
	}
}

// ---- helpers ----

func nullUUIDPtr(n uuid.NullUUID) *string {
	if n.Valid {
		s := n.UUID.String()
		return &s
	}
	return nil
}

func nullTimePtr(t sql.NullTime) *time.Time {
	if t.Valid {
		return &t.Time
	}
	return nil
}

func optUUID(s *string) uuid.NullUUID {
	if s == nil {
		return uuid.NullUUID{Valid: false}
	}
	return uuid.NullUUID{UUID: uuid.MustParse(*s), Valid: true}
}

func optTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: *t, Valid: true}
}
