package database

import (
	"context"

	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	"github.com/irosadie/open-state/api/internal/infrastructure/db"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

// PgxContextRepository implements IContextRepository on PostgreSQL via sqlc.
type PgxContextRepository struct {
	queries *db.Queries
}

// NewPgxContextRepository returns a PostgreSQL-backed IContextRepository.
func NewPgxContextRepository(pool *pgxpool.Pool) repositories.IContextRepository {
	sqlDB := stdlib.OpenDBFromPool(pool)
	return newPgxContextRepository(db.New(sqlDB))
}

// newPgxContextRepository builds an IContextRepository from a sqlc queries handle.
// It enables the composed PostgresAdapter to bind this repository to a shared
// transaction via WithTx.
func newPgxContextRepository(q *db.Queries) repositories.IContextRepository {
	return &PgxContextRepository{queries: q}
}

func (r *PgxContextRepository) UpsertContext(ctx context.Context, tenantID string, scopeType entities.ContextScopeType, scopeID, key string, value []byte, expectedVersion int) (*entities.ContextRecord, error) {
	row, err := r.queries.UpsertContext(ctx, db.UpsertContextParams{
		TenantID:  mustUUID(tenantID),
		ScopeType: string(scopeType),
		ScopeID:   scopeID,
		Key:       key,
		Value:     value,
		Version:   int32(expectedVersion),
	})
	if err != nil {
		return nil, mapOptimisticError(err)
	}
	return mapContextRecord(row), nil
}

func (r *PgxContextRepository) FindContextByScope(ctx context.Context, tenantID string, scopeType entities.ContextScopeType, scopeID, key string) (*entities.ContextRecord, error) {
	row, err := r.queries.FindContextByScope(ctx, db.FindContextByScopeParams{
		TenantID:  mustUUID(tenantID),
		ScopeType: string(scopeType),
		ScopeID:   scopeID,
		Key:       key,
	})
	if err != nil {
		return nil, mapNotFound(err, "context")
	}
	return mapContextRecord(row), nil
}

func (r *PgxContextRepository) ListContextByScope(ctx context.Context, tenantID string, scopeType entities.ContextScopeType, scopeID string) ([]entities.ContextRecord, error) {
	rows, err := r.queries.ListContextByScope(ctx, db.ListContextByScopeParams{
		TenantID:  mustUUID(tenantID),
		ScopeType: string(scopeType),
		ScopeID:   scopeID,
	})
	if err != nil {
		return nil, err
	}
	out := make([]entities.ContextRecord, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapContextRecord(row))
	}
	return out, nil
}

func (r *PgxContextRepository) DeleteContext(ctx context.Context, tenantID string, scopeType entities.ContextScopeType, scopeID, key string) error {
	return r.queries.DeleteContext(ctx, db.DeleteContextParams{
		TenantID:  mustUUID(tenantID),
		ScopeType: string(scopeType),
		ScopeID:   scopeID,
		Key:       key,
	})
}

func (r *PgxContextRepository) UpsertMemoryReference(ctx context.Context, tenantID, ownerType, ownerID, name string, value []byte, sourceWorkflowInstanceID *string) (*entities.MemoryReference, error) {
	row, err := r.queries.UpsertMemoryReference(ctx, db.UpsertMemoryReferenceParams{
		TenantID:                 mustUUID(tenantID),
		OwnerType:                ownerType,
		OwnerID:                  ownerID,
		Name:                     name,
		Value:                    value,
		SourceWorkflowInstanceID: optUUID(sourceWorkflowInstanceID),
	})
	if err != nil {
		return nil, mapPgError(err, "upsert memory reference")
	}
	return mapMemoryReference(row), nil
}

func (r *PgxContextRepository) FindMemoryReference(ctx context.Context, tenantID, ownerType, ownerID, name string) (*entities.MemoryReference, error) {
	row, err := r.queries.FindMemoryReference(ctx, db.FindMemoryReferenceParams{
		TenantID:  mustUUID(tenantID),
		OwnerType: ownerType,
		OwnerID:   ownerID,
		Name:      name,
	})
	if err != nil {
		return nil, mapNotFound(err, "memory reference")
	}
	return mapMemoryReference(row), nil
}

func (r *PgxContextRepository) ListMemoryByOwner(ctx context.Context, tenantID, ownerType, ownerID string) ([]entities.MemoryReference, error) {
	rows, err := r.queries.ListMemoryByOwner(ctx, db.ListMemoryByOwnerParams{
		TenantID:  mustUUID(tenantID),
		OwnerType: ownerType,
		OwnerID:   ownerID,
	})
	if err != nil {
		return nil, err
	}
	out := make([]entities.MemoryReference, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapMemoryReference(row))
	}
	return out, nil
}

func (r *PgxContextRepository) DeleteMemoryReference(ctx context.Context, tenantID, ownerType, ownerID, name string) error {
	return r.queries.DeleteMemoryReference(ctx, db.DeleteMemoryReferenceParams{
		TenantID:  mustUUID(tenantID),
		OwnerType: ownerType,
		OwnerID:   ownerID,
		Name:      name,
	})
}

// ---- mappers ----

func mapContextRecord(row db.ContextRecord) *entities.ContextRecord {
	return &entities.ContextRecord{
		ID:        row.ID.String(),
		TenantID:  row.TenantID.String(),
		ScopeType: entities.ContextScopeType(row.ScopeType),
		ScopeID:   row.ScopeID,
		Key:       row.Key,
		Value:     row.Value,
		Version:   int(row.Version),
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

func mapMemoryReference(row db.MemoryReference) *entities.MemoryReference {
	return &entities.MemoryReference{
		ID:                       row.ID.String(),
		TenantID:                 row.TenantID.String(),
		OwnerType:                row.OwnerType,
		OwnerID:                  row.OwnerID,
		Name:                     row.Name,
		Value:                    row.Value,
		SourceWorkflowInstanceID: nullUUIDPtr(row.SourceWorkflowInstanceID),
		CreatedAt:                row.CreatedAt,
		UpdatedAt:                row.UpdatedAt,
	}
}

var _ repositories.IContextRepository = (*PgxContextRepository)(nil)
