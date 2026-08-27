package database

import (
	"context"

	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	"github.com/irosadie/open-state/api/internal/infrastructure/db"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

// PgxProjectRepository implements IProjectRepository on PostgreSQL via sqlc.
type PgxProjectRepository struct {
	queries *db.Queries
}

// NewPgxProjectRepository returns a PostgreSQL-backed IProjectRepository.
func NewPgxProjectRepository(pool *pgxpool.Pool) repositories.IProjectRepository {
	sqlDB := stdlib.OpenDBFromPool(pool)
	return &PgxProjectRepository{queries: db.New(sqlDB)}
}

func (r *PgxProjectRepository) Create(ctx context.Context, tenantID, name, slug string, status entities.ProjectStatus) (*entities.Project, error) {
	row, err := r.queries.CreateProject(ctx, db.CreateProjectParams{
		TenantID: mustUUID(tenantID),
		Name:     name,
		Slug:     slug,
		Status:   string(status),
	})
	if err != nil {
		return nil, mapPgError(err, "create project")
	}
	return mapProject(row), nil
}

func (r *PgxProjectRepository) FindByID(ctx context.Context, tenantID, id string) (*entities.Project, error) {
	row, err := r.queries.FindProjectByID(ctx, db.FindProjectByIDParams{
		ID:       mustUUID(id),
		TenantID: mustUUID(tenantID),
	})
	if err != nil {
		return nil, mapNotFound(err, "project")
	}
	return mapProject(row), nil
}

func (r *PgxProjectRepository) FindBySlug(ctx context.Context, tenantID, slug string) (*entities.Project, error) {
	row, err := r.queries.FindProjectBySlug(ctx, db.FindProjectBySlugParams{
		Slug:     slug,
		TenantID: mustUUID(tenantID),
	})
	if err != nil {
		return nil, mapNotFound(err, "project")
	}
	return mapProject(row), nil
}

func (r *PgxProjectRepository) ListByTenant(ctx context.Context, tenantID string) ([]entities.Project, error) {
	rows, err := r.queries.ListProjectsByTenant(ctx, mustUUID(tenantID))
	if err != nil {
		return nil, err
	}
	out := make([]entities.Project, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapProject(row))
	}
	return out, nil
}

func mapProject(row db.Project) *entities.Project {
	return &entities.Project{
		ID:        row.ID.String(),
		TenantID:  row.TenantID.String(),
		Name:      row.Name,
		Slug:      row.Slug,
		Status:    entities.ProjectStatus(row.Status),
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}
