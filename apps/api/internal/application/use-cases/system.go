package usecases

import (
	"context"
	"os"

	"github.com/irosadie/open-state/api/internal/application/dtos"
)

type GetHealthUseCase struct{}

func NewGetHealthUseCase() *GetHealthUseCase {
	return &GetHealthUseCase{}
}

func (uc *GetHealthUseCase) Execute(_ context.Context) *dtos.HealthDTO {
	return &dtos.HealthDTO{Status: "ok"}
}

type GetAppInfoUseCase struct{}

func NewGetAppInfoUseCase() *GetAppInfoUseCase {
	return &GetAppInfoUseCase{}
}

func (uc *GetAppInfoUseCase) Execute(_ context.Context) *dtos.AppInfoDTO {
	env := os.Getenv("NODE_ENV")
	if env == "" {
		env = "development"
	}
	return &dtos.AppInfoDTO{
		Name:    "openstate",
		Version: "0.1.0",
		Env:     env,
	}
}
