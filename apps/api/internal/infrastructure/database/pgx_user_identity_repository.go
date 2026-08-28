package database

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	"github.com/irosadie/open-state/api/internal/infrastructure/db"
)

// PgxUserIdentityRepository implements IUserIdentityRepository on PostgreSQL.
type PgxUserIdentityRepository struct {
	queries *db.Queries
}

// NewPgxUserIdentityRepository returns a PostgreSQL-backed IUserIdentityRepository.
func NewPgxUserIdentityRepository(pool *pgxpool.Pool) repositories.IUserIdentityRepository {
	sqlDB := stdlib.OpenDBFromPool(pool)
	return newPgxUserIdentityRepository(db.New(sqlDB))
}

// newPgxUserIdentityRepository builds from a sqlc queries handle (for WithTx).
func newPgxUserIdentityRepository(q *db.Queries) repositories.IUserIdentityRepository {
	return &PgxUserIdentityRepository{queries: q}
}

func (r *PgxUserIdentityRepository) FindByProviderSubject(ctx context.Context, provider, subjectID string) (*entities.UserIdentity, error) {
	row, err := r.queries.FindIdentityByProviderSubject(ctx, db.FindIdentityByProviderSubjectParams{
		Provider:  provider,
		SubjectID: subjectID,
	})
	if err != nil {
		if isNoRows(err) {
			return nil, nil
		}
		return nil, mapPgError(err, "find identity by provider subject")
	}
	return mapUserIdentity(row), nil
}

func (r *PgxUserIdentityRepository) Create(ctx context.Context, userID, provider, subjectID string, autoProvisioned bool) (*entities.UserIdentity, error) {
	row, err := r.queries.CreateIdentity(ctx, db.CreateIdentityParams{
		UserID:          mustUUID(userID),
		Provider:        provider,
		SubjectID:       subjectID,
		AutoProvisioned: autoProvisioned,
	})
	if err != nil {
		return nil, mapPgError(err, "create identity")
	}
	return mapUserIdentity(row), nil
}

func (r *PgxUserIdentityRepository) ListByUser(ctx context.Context, userID string) ([]entities.UserIdentity, error) {
	rows, err := r.queries.ListIdentitiesByUser(ctx, mustUUID(userID))
	if err != nil {
		return nil, err
	}
	out := make([]entities.UserIdentity, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapUserIdentity(row))
	}
	return out, nil
}

func mapUserIdentity(row db.UserIdentity) *entities.UserIdentity {
	return &entities.UserIdentity{
		ID:              row.ID.String(),
		UserID:          row.UserID.String(),
		Provider:        row.Provider,
		SubjectID:       row.SubjectID,
		AutoProvisioned: row.AutoProvisioned,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
}

func isNoRows(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}
