package usecases

import (
	"context"

	domain "github.com/vibecoding-starter/go-shared/domain"
	"github.com/vibecoding-starter/api/internal/domain/entities"
	"github.com/vibecoding-starter/api/internal/domain/repositories"
)

type GetCurrentUserUseCase struct {
	repo repositories.IAuthRepository
}

func NewGetCurrentUserUseCase(repo repositories.IAuthRepository) *GetCurrentUserUseCase {
	return &GetCurrentUserUseCase{repo: repo}
}

func (uc *GetCurrentUserUseCase) Execute(ctx context.Context, userID string) (*entities.User, error) {
	user, err := uc.repo.FindUserByID(ctx, userID)
	if err != nil || user == nil {
		return nil, domain.NewNotFound("user not found")
	}
	return user, nil
}
