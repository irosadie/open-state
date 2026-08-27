package services

import (
	"context"

	"github.com/vibecoding-starter/api/internal/application/dtos"
	usecases "github.com/vibecoding-starter/api/internal/application/use-cases"
	"github.com/vibecoding-starter/api/internal/domain/entities"
)

type AuthService struct {
	registerUC *usecases.RegisterUserUseCase
	loginUC    *usecases.LoginUserUseCase
	logoutUC   *usecases.LogoutUserUseCase
	getMeUC    *usecases.GetCurrentUserUseCase
}

func NewAuthService(
	registerUC *usecases.RegisterUserUseCase,
	loginUC *usecases.LoginUserUseCase,
	logoutUC *usecases.LogoutUserUseCase,
	getMeUC *usecases.GetCurrentUserUseCase,
) *AuthService {
	return &AuthService{
		registerUC: registerUC,
		loginUC:    loginUC,
		logoutUC:   logoutUC,
		getMeUC:    getMeUC,
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
