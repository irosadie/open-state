package controllers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	usecases "github.com/irosadie/open-state/api/internal/application/use-cases"
)

type SystemController struct {
	healthUC  *usecases.GetHealthUseCase
	appInfoUC *usecases.GetAppInfoUseCase
}

func NewSystemController(healthUC *usecases.GetHealthUseCase, appInfoUC *usecases.GetAppInfoUseCase) *SystemController {
	return &SystemController{healthUC: healthUC, appInfoUC: appInfoUC}
}

func (ctrl *SystemController) Health(c echo.Context) error {
	return c.JSON(http.StatusOK, ctrl.healthUC.Execute(c.Request().Context()))
}

func (ctrl *SystemController) AppInfo(c echo.Context) error {
	return c.JSON(http.StatusOK, ctrl.appInfoUC.Execute(c.Request().Context()))
}
