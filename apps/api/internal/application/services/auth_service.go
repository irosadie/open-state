package services

import (
	"context"

	"github.com/irosadie/open-state/api/internal/application/dtos"
	usecases "github.com/irosadie/open-state/api/internal/application/use-cases"
	"github.com/irosadie/open-state/api/internal/domain/entities"
)

type AuthService struct {
	registerUC *usecases.RegisterUserUseCase
	loginUC    *usecases.LoginUserUseCase
	logoutUC   *usecases.LogoutUserUseCase
	getMeUC    *usecases.GetCurrentUserUseCase
	authz      *AuthorizationService
}

func NewAuthService(
	registerUC *usecases.RegisterUserUseCase,
	loginUC *usecases.LoginUserUseCase,
	logoutUC *usecases.LogoutUserUseCase,
	getMeUC *usecases.GetCurrentUserUseCase,
	authz *AuthorizationService,
) *AuthService {
	return &AuthService{
		registerUC: registerUC,
		loginUC:    loginUC,
		logoutUC:   logoutUC,
		getMeUC:    getMeUC,
		authz:      authz,
	}
}

func (s *AuthService) Register(ctx context.Context, input usecases.RegisterUserInput) (*dtos.UserDTO, error) {
	user, err := s.registerUC.Execute(ctx, input)
	if err != nil {
		return nil, err
	}
	return toUserDTO(user), nil
}

func (s *AuthService) Login(ctx context.Context, input usecases.LoginUserInput) (*dtos.LoginDTO, error) {
	out, err := s.loginUC.Execute(ctx, input)
	if err != nil {
		return nil, err
	}
	return &dtos.LoginDTO{
		AccessToken: out.AccessToken,
		User:        *toUserDTO(out.User),
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, accessToken string) error {
	return s.logoutUC.Execute(ctx, accessToken)
}

func (s *AuthService) GetCurrentUser(ctx context.Context, userID string) (*dtos.UserDTO, error) {
	user, err := s.getMeUC.Execute(ctx, userID)
	if err != nil {
		return nil, err
	}
	return toUserDTO(user), nil
}

// GetCurrentUserForTenant returns the current user enriched with their effective
// tenant role and granted permissions (PRD 80, 81). The role and permissions are
// derived from role_assignments for the given tenant; an absent assignment
// resolves to the least-privilege VIEWER.
func (s *AuthService) GetCurrentUserForTenant(ctx context.Context, userID, tenantID string) (*dtos.UserDTO, error) {
	dto, err := s.GetCurrentUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	role, err := s.authz.RoleForTenant(ctx, userID, tenantID)
	if err != nil {
		return nil, err
	}
	perms, err := s.authz.PermissionsForTenant(ctx, userID, tenantID)
	if err != nil {
		return nil, err
	}
	dto.Role = string(role)
	dto.Permissions = make([]string, 0, len(perms))
	for _, p := range perms {
		dto.Permissions = append(dto.Permissions, string(p))
	}
	return dto, nil
}

func toUserDTO(u *entities.User) *dtos.UserDTO {
	return &dtos.UserDTO{
		ID:     u.ID,
		Email:  u.Email,
		Name:   u.Name,
		Role:   string(u.Role),
		Status: string(u.Status),
		Photo:  u.Photo,
	}
}
