package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/vibecoding-starter/api/internal/domain/entities"
	"github.com/vibecoding-starter/api/internal/domain/repositories"
	"github.com/vibecoding-starter/api/internal/infrastructure/db"
)

type PgxAuthRepository struct {
	queries *db.Queries
}

func NewPgxAuthRepository(pool *pgxpool.Pool) repositories.IAuthRepository {
	sqlDB := stdlib.OpenDBFromPool(pool)
	return &PgxAuthRepository{queries: db.New(sqlDB)}
}

func (r *PgxAuthRepository) FindUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	row, err := r.queries.FindUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return mapUser(row), nil
}

func (r *PgxAuthRepository) FindUserByID(ctx context.Context, id string) (*entities.User, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	row, err := r.queries.FindUserByID(ctx, uid)
	if err != nil {
		return nil, err
	}
	return mapUser(row), nil
}

func (r *PgxAuthRepository) CreateUser(ctx context.Context, email, passwordHash, name string, role entities.UserRole, status entities.UserStatus) (*entities.User, error) {
	row, err := r.queries.CreateUser(ctx, db.CreateUserParams{
		Email:        email,
		PasswordHash: passwordHash,
		Name:         name,
		Role:         db.UserRole(role),
		Status:       db.UserStatus(status),
	})
	if err != nil {
		return nil, err
	}
	return mapUser(row), nil
}

func (r *PgxAuthRepository) CreateSession(ctx context.Context, userID, tokenHash string, expiresAt time.Time) (*entities.AuthSession, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	row, err := r.queries.CreateAuthSession(ctx, db.CreateAuthSessionParams{
		UserID:    uid,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return nil, err
	}
	return &entities.AuthSession{
		ID:        row.ID.String(),
		UserID:    row.UserID.String(),
		TokenHash: row.TokenHash,
		ExpiresAt: row.ExpiresAt,
		CreatedAt: row.CreatedAt,
	}, nil
}

func (r *PgxAuthRepository) FindSessionByTokenHash(ctx context.Context, tokenHash string) (*entities.AuthSession, error) {
	row, err := r.queries.FindSessionByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, err
	}
	return &entities.AuthSession{
		ID:        row.ID.String(),
		UserID:    row.UserID.String(),
		TokenHash: row.TokenHash,
		ExpiresAt: row.ExpiresAt,
		CreatedAt: row.CreatedAt,
	}, nil
}

func (r *PgxAuthRepository) DeleteSessionByTokenHash(ctx context.Context, tokenHash string) error {
	return r.queries.DeleteSessionByTokenHash(ctx, tokenHash)
}

func mapUser(row db.User) *entities.User {
	u := &entities.User{
		ID:           row.ID.String(),
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		Name:         row.Name,
		Role:         entities.UserRole(row.Role),
		Status:       entities.UserStatus(row.Status),
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
	if row.Photo.Valid {
		photo := row.Photo.String
		u.Photo = &photo
	}
	return u
}

// mapUser also handles db.FindUserByEmailRow and db.FindUserByIDRow since sqlc
// generates separate row types per query — use a shared helper via sql.NullString.
func nullStringPtr(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}
