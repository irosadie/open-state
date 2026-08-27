package usecases

import (
	"context"

	domain "github.com/irosadie/open-state/go-shared/domain"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	"github.com/irosadie/open-state/api/internal/domain/services"
	"golang.org/x/crypto/bcrypt"
)

type RegisterUserInput struct {
	Email    string
	Password string
	Name     string
}

type RegisterUserUseCase struct {
	repo  repositories.IAuthRepository
	token services.TokenService
}

func NewRegisterUserUseCase(repo repositories.IAuthRepository, token services.TokenService) *RegisterUserUseCase {
	return &RegisterUserUseCase{repo: repo, token: token}
}

func (uc *RegisterUserUseCase) Execute(ctx context.Context, input RegisterUserInput) (*entities.User, error) {
	existing, _ := uc.repo.FindUserByEmail(ctx, input.Email)
	if existing != nil {
		return nil, domain.NewConflict("email already in use")
	}

	if input.Email == "" || input.Password == "" || input.Name == "" {
		return nil, domain.NewValidation("email, password, and name are required")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, domain.NewInternal("failed to hash password")
	}

	user, err := uc.repo.CreateUser(ctx, input.Email, string(hash), input.Name, entities.UserRoleUser, entities.UserStatusActive)
	if err != nil {
		return nil, domain.NewInternal("failed to create user")
	}

	return user, nil
}
