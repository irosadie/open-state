package database

import (
	"context"

	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	"github.com/irosadie/open-state/api/internal/infrastructure/db"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

// PgxRoleAssignmentRepository implements IRoleAssignmentRepository on PostgreSQL
// via sqlc (PRD 80, 81). Every query is tenant-scoped by an explicit tenantID.
type PgxRoleAssignmentRepository struct {
	queries *db.Queries
}

// NewPgxRoleAssignmentRepository returns a PostgreSQL-backed IRoleAssignmentRepository.
func NewPgxRoleAssignmentRepository(pool *pgxpool.Pool) repositories.IRoleAssignmentRepository {
	sqlDB := stdlib.OpenDBFromPool(pool)
	return newPgxRoleAssignmentRepository(db.New(sqlDB))
}

// newPgxRoleAssignmentRepository builds an IRoleAssignmentRepository from a sqlc
// queries handle, enabling composition via the PostgresAdapter.WithTx helper.
func newPgxRoleAssignmentRepository(q *db.Queries) repositories.IRoleAssignmentRepository {
	return &PgxRoleAssignmentRepository{queries: q}
}

func (r *PgxRoleAssignmentRepository) Assign(ctx context.Context, userID, tenantID string, role entities.UserRole) (*entities.RoleAssignment, error) {
	row, err := r.queries.AssignRole(ctx, db.AssignRoleParams{
		UserID:   mustUUID(userID),
		TenantID: mustUUID(tenantID),
		Role:     string(role),
	})
	if err != nil {
		return nil, mapPgError(err, "assign role")
	}
	return mapRoleAssignment(row), nil
}

func (r *PgxRoleAssignmentRepository) FindRoleByUserAndTenant(ctx context.Context, userID, tenantID string) (entities.UserRole, error) {
	row, err := r.queries.FindRoleByUserAndTenant(ctx, db.FindRoleByUserAndTenantParams{
		UserID:   mustUUID(userID),
		TenantID: mustUUID(tenantID),
	})
	if err != nil {
		return "", mapPgError(err, "find role by user and tenant")
	}
	return entities.UserRole(row.Role), nil
}

func (r *PgxRoleAssignmentRepository) ListByTenant(ctx context.Context, tenantID string) ([]entities.RoleAssignment, error) {
	rows, err := r.queries.ListRolesByTenant(ctx, mustUUID(tenantID))
	if err != nil {
		return nil, err
	}
	out := make([]entities.RoleAssignment, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapRoleAssignment(row))
	}
	return out, nil
}

func (r *PgxRoleAssignmentRepository) Remove(ctx context.Context, userID, tenantID string) error {
	return r.queries.RemoveRoleAssignment(ctx, db.RemoveRoleAssignmentParams{
		UserID:   mustUUID(userID),
		TenantID: mustUUID(tenantID),
	})
}

func mapRoleAssignment(row db.RoleAssignment) *entities.RoleAssignment {
	return &entities.RoleAssignment{
		ID:        row.ID.String(),
		UserID:    row.UserID.String(),
		TenantID:  row.TenantID.String(),
		Role:      entities.UserRole(row.Role),
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}
