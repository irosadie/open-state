package usecases

import (
	"context"

	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	"github.com/irosadie/open-state/api/internal/domain/services"
	domain "github.com/irosadie/open-state/go-shared/domain"
	"golang.org/x/crypto/bcrypt"
)

type RegisterUserInput struct {
	Email    string
	Password string
	Name     string
}

type RegisterUserUseCase struct {
	repo  repositories.IAuthRepository
	roles roleAssigner
	token services.TokenService
}

type roleAssigner interface {
	Assign(ctx context.Context, userID, tenantID string, role entities.UserRole) (*entities.RoleAssignment, error)
}

const defaultTenantID = "00000000-0000-0000-0000-000000000001"

func NewRegisterUserUseCase(repo repositories.IAuthRepository, roles roleAssigner, token services.TokenService) *RegisterUserUseCase {
	return &RegisterUserUseCase{repo: repo, roles: roles, token: token}
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

	// users.role is a deprecated legacy column (ENUM 'USER'/'ADMIN'). The
	// effective role is tenant-scoped and lives in role_assignments (PRD 81),
	// defaulting to least-privilege VIEWER when absent. Write the legacy value
	// here only to satisfy the deprecated column.
	user, err := uc.repo.CreateUser(ctx, input.Email, string(hash), input.Name, entities.UserRoleLegacy, entities.UserStatusActive)
	if err != nil {
		return nil, domain.NewInternal("failed to create user")
	}
	if _, err := uc.roles.Assign(ctx, user.ID, defaultTenantID, entities.UserRoleViewer); err != nil {
		return nil, domain.NewInternal("failed to assign default role")
	}

	return user, nil
}
