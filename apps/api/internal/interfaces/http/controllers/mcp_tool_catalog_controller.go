package controllers

import (
	"net/http"

	"github.com/irosadie/open-state/api/internal/application/dtos"
	appservices "github.com/irosadie/open-state/api/internal/application/services"
	"github.com/irosadie/open-state/api/internal/interfaces/http/middleware"
	"github.com/irosadie/open-state/go-shared/domain"
	"github.com/labstack/echo/v4"
)

// MCPToolCatalogController exposes explicit, project-scoped MCP tools/list
// discovery and local enablement controls. It has no provider invocation route.
type MCPToolCatalogController struct {
	svc *appservices.MCPToolCatalogService
}

func NewMCPToolCatalogController(svc *appservices.MCPToolCatalogService) *MCPToolCatalogController {
	return &MCPToolCatalogController{svc: svc}
}

func (ctrl *MCPToolCatalogController) List(c echo.Context) error {
	tenantID, projectID, err := mcpScope(c)
	if err != nil {
		return err
	}
	result, err := ctrl.svc.List(c.Request().Context(), tenantID, projectID, c.Param("id"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"data": result})
}

func (ctrl *MCPToolCatalogController) Refresh(c echo.Context) error {
	tenantID, projectID, err := mcpScope(c)
	if err != nil {
		return err
	}
	actor, _ := c.Get(middleware.UserIDKey).(string)
	result, err := ctrl.svc.Refresh(c.Request().Context(), tenantID, projectID, c.Param("id"), actor)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"data": result})
}

func (ctrl *MCPToolCatalogController) SetEnabled(c echo.Context) error {
	tenantID, projectID, err := mcpScope(c)
	if err != nil {
		return err
	}
	var req dtos.SetMCPToolEnabledRequest
	if err := c.Bind(&req); err != nil || req.Enabled == nil {
		return domain.NewValidation("enabled is required")
	}
	actor, _ := c.Get(middleware.UserIDKey).(string)
	result, err := ctrl.svc.SetEnabled(c.Request().Context(), tenantID, projectID, c.Param("id"), c.Param("toolName"), actor, *req.Enabled)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"data": result})
}
