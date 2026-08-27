package usecases

import (
	"context"
	"time"

	domain "github.com/irosadie/open-state/go-shared/domain"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	"github.com/irosadie/open-state/api/internal/domain/services"
	"golang.org/x/crypto/bcrypt"
)

type LoginUserInput struct {
	Email    string
	Password string
}

type LoginUserOutput struct {
	User        *entities.User
	AccessToken string
}

type LoginUserUseCase struct {
	repo  repositories.IAuthRepository
	token services.TokenService
}

func NewLoginUserUseCase(repo repositories.IAuthRepository, token services.TokenService) *LoginUserUseCase {
	return &LoginUserUseCase{repo: repo, token: token}
}

func (uc *LoginUserUseCase) Execute(ctx context.Context, input LoginUserInput) (*LoginUserOutput, error) {
	user, err := uc.repo.FindUserByEmail(ctx, input.Email)
	if err != nil || user == nil {
		return nil, domain.NewUnauthorized("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, domain.NewUnauthorized("invalid credentials")
	}

	accessToken, err := uc.token.GenerateToken(user.ID)
	if err != nil {
		return nil, domain.NewInternal("failed to generate token")
	}

	tokenHash := uc.token.HashToken(accessToken)
	expiresAt := time.Now().Add(24 * time.Hour)

	_, err = uc.repo.CreateSession(ctx, user.ID, tokenHash, expiresAt)
	if err != nil {
		return nil, domain.NewInternal("failed to create session")
	}

	return &LoginUserOutput{User: user, AccessToken: accessToken}, nil
}
