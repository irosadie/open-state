package controllers

import (
	"net/http"

	"github.com/irosadie/open-state/api/internal/application/dtos"
	appservices "github.com/irosadie/open-state/api/internal/application/services"
	"github.com/irosadie/open-state/api/internal/interfaces/http/middleware"
	domain "github.com/irosadie/open-state/go-shared/domain"
	"github.com/labstack/echo/v4"
)

// ProjectCapabilityMCPBindingController exposes the safe project-scoped
// catalog and logical capability binding APIs used by State Builder.
type ProjectCapabilityMCPBindingController struct {
	svc *appservices.ProjectCapabilityMCPBindingService
}

func NewProjectCapabilityMCPBindingController(svc *appservices.ProjectCapabilityMCPBindingService) *ProjectCapabilityMCPBindingController {
	return &ProjectCapabilityMCPBindingController{svc: svc}
}

func (ctrl *ProjectCapabilityMCPBindingController) ListOptions(c echo.Context) error {
	tenantID, projectID, err := mcpScope(c)
	if err != nil {
		return err
	}
	result, err := ctrl.svc.ListOptions(c.Request().Context(), tenantID, projectID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, result)
}

func (ctrl *ProjectCapabilityMCPBindingController) List(c echo.Context) error {
	tenantID, projectID, err := mcpScope(c)
	if err != nil {
		return err
	}
	result, err := ctrl.svc.List(c.Request().Context(), tenantID, projectID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, result)
}

func (ctrl *ProjectCapabilityMCPBindingController) Upsert(c echo.Context) error {
	tenantID, projectID, err := mcpScope(c)
	if err != nil {
		return err
	}
	var req dtos.UpsertProjectCapabilityMCPBindingRequest
	if err := c.Bind(&req); err != nil {
		return domain.NewValidation("invalid request body")
	}
	actor, _ := c.Get(middleware.UserIDKey).(string)
	result, err := ctrl.svc.Upsert(c.Request().Context(), tenantID, projectID, c.Param("capabilityId"), actor, req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"data": result})
}

func (ctrl *ProjectCapabilityMCPBindingController) Delete(c echo.Context) error {
	tenantID, projectID, err := mcpScope(c)
	if err != nil {
		return err
	}
	actor, _ := c.Get(middleware.UserIDKey).(string)
	if err := ctrl.svc.Delete(c.Request().Context(), tenantID, projectID, c.Param("capabilityId"), actor); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "project MCP capability binding deleted"})
}
