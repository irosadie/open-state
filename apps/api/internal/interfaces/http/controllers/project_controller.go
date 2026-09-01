package controllers

import (
	"net/http"

	appservices "github.com/irosadie/open-state/api/internal/application/services"
	"github.com/labstack/echo/v4"
)

// ProjectController exposes tenant-scoped project discovery for authenticated
// clients that need to select existing project IDs.
type ProjectController struct {
	svc *appservices.ProjectService
}

// NewProjectController builds a ProjectController.
func NewProjectController(svc *appservices.ProjectService) *ProjectController {
	return &ProjectController{svc: svc}
}

// List returns all projects belonging to the tenant selected by the request
// header. The shared API client unwraps the data envelope for the frontend.
func (ctrl *ProjectController) List(c echo.Context) error {
	tenantID, err := tenantFromHeader(c)
	if err != nil {
		return err
	}
	result, err := ctrl.svc.List(c.Request().Context(), tenantID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, result)
}
