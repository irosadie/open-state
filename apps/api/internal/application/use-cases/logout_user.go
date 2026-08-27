package usecases

import (
	"context"

	domain "github.com/vibecoding-starter/go-shared/domain"
	"github.com/vibecoding-starter/api/internal/domain/repositories"
	"github.com/vibecoding-starter/api/internal/domain/services"
)

type LogoutUserUseCase struct {
	repo  repositories.IAuthRepository
	token services.TokenService
}

func NewLogoutUserUseCase(repo repositories.IAuthRepository, token services.TokenService) *LogoutUserUseCase {
	return &LogoutUserUseCase{repo: repo, token: token}
}

func (uc *LogoutUserUseCase) Execute(ctx context.Context, accessToken string) error {
	tokenHash := uc.token.HashToken(accessToken)

	session, err := uc.repo.FindSessionByTokenHash(ctx, tokenHash)
	if err != nil || session == nil {
		return domain.NewUnauthorized("session not found")
	}

	if err := uc.repo.DeleteSessionByTokenHash(ctx, tokenHash); err != nil {
		return domain.NewInternal("failed to delete session")
	}

	return nil
}
