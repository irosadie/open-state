package database

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	"github.com/irosadie/open-state/api/internal/infrastructure/db"
	"github.com/irosadie/open-state/go-shared/domain"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

// PgxAdminRepository implements tenant profile and membership persistence via
// sqlc. The transaction handle is retained so role changes can check and write
// the last-Owner invariant atomically.
type PgxAdminRepository struct {
	queries *db.Queries
	db      *sql.DB
	tx      *sql.Tx
}

func NewPgxAdminRepository(pool *pgxpool.Pool) repositories.IAdminRepository {
	sqlDB := stdlib.OpenDBFromPool(pool)
	return newPgxAdminRepository(db.New(sqlDB), sqlDB)
}

func newPgxAdminRepository(queries *db.Queries, sqlDB *sql.DB) *PgxAdminRepository {
	return &PgxAdminRepository{queries: queries, db: sqlDB}
}

func (r *PgxAdminRepository) FindTenantByID(ctx context.Context, tenantID string) (*entities.Tenant, error) {
	id, err := parseAdminUUID(tenantID)
	if err != nil {
		return nil, err
	}
	row, err := r.queries.FindTenantByID(ctx, id)
	if err != nil {
		return nil, mapNotFound(err, "tenant")
	}
	return mapTenant(row), nil
}

func (r *PgxAdminRepository) UpdateTenantProfile(ctx context.Context, tenantID, name, slug, description string) (*entities.Tenant, error) {
	id, err := parseAdminUUID(tenantID)
	if err != nil {
		return nil, err
	}
	row, err := r.queries.UpdateTenantProfile(ctx, db.UpdateTenantProfileParams{
		ID:          id,
		Name:        name,
		Slug:        slug,
		Description: description,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.NewNotFound("tenant not found")
		}
		return nil, mapPgError(err, "update tenant profile")
	}
	return mapTenant(row), nil
}

func (r *PgxAdminRepository) ListMemberships(ctx context.Context, tenantID string, search *string, offset, limit int) ([]entities.TenantMembership, error) {
	id, err := parseAdminUUID(tenantID)
	if err != nil {
		return nil, err
	}
	rows, err := r.queries.ListAdminMemberships(ctx, db.ListAdminMembershipsParams{
		TenantID:   id,
		Search:     nullString(search),
		PageOffset: int32(offset),
		PageSize:   int32(limit),
	})
	if err != nil {
		return nil, err
	}
	out := make([]entities.TenantMembership, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapListMembership(row))
	}
	return out, nil
}

func (r *PgxAdminRepository) CountMemberships(ctx context.Context, tenantID string, search *string) (int64, error) {
	id, err := parseAdminUUID(tenantID)
	if err != nil {
		return 0, err
	}
	return r.queries.CountAdminMemberships(ctx, db.CountAdminMembershipsParams{
		TenantID: id,
		Search:   nullString(search),
	})
}

func (r *PgxAdminRepository) FindMembership(ctx context.Context, tenantID, userID string) (*entities.TenantMembership, error) {
	tid, err := parseAdminUUID(tenantID)
	if err != nil {
		return nil, err
	}
	uid, err := parseAdminUUID(userID)
	if err != nil {
		return nil, err
	}
	row, err := r.queries.FindAdminMembership(ctx, db.FindAdminMembershipParams{
		TenantID: tid,
		UserID:   uid,
	})
	if err != nil {
		return nil, mapNotFound(err, "membership")
	}
	membership := mapFindMembership(row)
	return &membership, nil
}

func (r *PgxAdminRepository) CountOwners(ctx context.Context, tenantID string) (int64, error) {
	id, err := parseAdminUUID(tenantID)
	if err != nil {
		return 0, err
	}
	return r.queries.CountAdminOwners(ctx, id)
}

func (r *PgxAdminRepository) AssignMembershipRole(ctx context.Context, tenantID, userID string, role entities.UserRole) (*entities.TenantMembership, error) {
	tid, err := parseAdminUUID(tenantID)
	if err != nil {
		return nil, err
	}
	uid, err := parseAdminUUID(userID)
	if err != nil {
		return nil, err
	}
	_, err = r.queries.UpsertAdminMembershipRole(ctx, db.UpsertAdminMembershipRoleParams{
		UserID:   uid,
		TenantID: tid,
		Role:     string(role),
	})
	if err != nil {
		return nil, mapPgError(err, "assign membership role")
	}
	return r.FindMembership(ctx, tenantID, userID)
}

func (r *PgxAdminRepository) RemoveMembership(ctx context.Context, tenantID, userID string) error {
	tid, err := parseAdminUUID(tenantID)
	if err != nil {
		return err
	}
	uid, err := parseAdminUUID(userID)
	if err != nil {
		return err
	}
	if err := r.queries.RemoveAdminMembership(ctx, db.RemoveAdminMembershipParams{
		TenantID: tid,
		UserID:   uid,
	}); err != nil {
		return mapPgError(err, "remove membership")
	}
	return nil
}

func (r *PgxAdminRepository) WithTx(ctx context.Context, fn func(repositories.IAdminRepository) error) error {
	if r.tx != nil {
		return fn(r)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.NewInternal(err.Error())
	}
	defer tx.Rollback() //nolint:errcheck // rollback is a no-op after commit

	txRepo := &PgxAdminRepository{
		queries: r.queries.WithTx(tx),
		db:      r.db,
		tx:      tx,
	}
	if err := fn(txRepo); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return domain.NewInternal(err.Error())
	}
	return nil
}

func mapTenant(row db.Tenant) *entities.Tenant {
	return &entities.Tenant{
		ID:          row.ID.String(),
		Name:        row.Name,
		Slug:        row.Slug,
		Description: row.Description,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

func mapListMembership(row db.ListAdminMembershipsRow) entities.TenantMembership {
	return entities.TenantMembership{
		RoleAssignmentID: row.RoleAssignmentID.String(),
		UserID:           row.UserID.String(),
		TenantID:         row.TenantID.String(),
		Role:             entities.UserRole(row.Role),
		Email:            row.Email,
		Name:             row.Name,
		Status:           entities.UserStatus(row.Status),
		Photo:            nullStringPtr(row.Photo),
		CreatedAt:        row.RoleCreatedAt,
		UpdatedAt:        row.RoleUpdatedAt,
	}
}

func mapFindMembership(row db.FindAdminMembershipRow) entities.TenantMembership {
	return entities.TenantMembership{
		RoleAssignmentID: row.RoleAssignmentID.String(),
		UserID:           row.UserID.String(),
		TenantID:         row.TenantID.String(),
		Role:             entities.UserRole(row.Role),
		Email:            row.Email,
		Name:             row.Name,
		Status:           entities.UserStatus(row.Status),
		Photo:            nullStringPtr(row.Photo),
		CreatedAt:        row.RoleCreatedAt,
		UpdatedAt:        row.RoleUpdatedAt,
	}
}

func parseAdminUUID(value string) (uuid.UUID, error) {
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, domain.NewValidation("invalid tenant or user id")
	}
	return id, nil
}
