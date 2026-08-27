package usecases

import (
	"context"
	"os"

	"github.com/vibecoding-starter/api/internal/application/dtos"
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
		Name:    "vibecoding-starter",
		Version: "0.1.0",
		Env:     env,
	}
}
