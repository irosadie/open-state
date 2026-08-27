package repositories

import (
	"context"
	"time"

	"github.com/irosadie/open-state/api/internal/domain/entities"
)

// IAuthRepository defines the persistence contract for auth domain.
type IAuthRepository interface {
	FindUserByEmail(ctx context.Context, email string) (*entities.User, error)
	FindUserByID(ctx context.Context, id string) (*entities.User, error)
	CreateUser(ctx context.Context, email, passwordHash, name string, role entities.UserRole, status entities.UserStatus) (*entities.User, error)
	CreateSession(ctx context.Context, userID, tokenHash string, expiresAt time.Time) (*entities.AuthSession, error)
	FindSessionByTokenHash(ctx context.Context, tokenHash string) (*entities.AuthSession, error)
	DeleteSessionByTokenHash(ctx context.Context, tokenHash string) error
}
