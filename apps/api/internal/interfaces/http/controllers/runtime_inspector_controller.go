package controllers

import (
	"net/http"

	"github.com/irosadie/open-state/api/internal/application/services"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/labstack/echo/v4"
)

// RuntimeInspectorController exposes the authenticated Runtime Inspector read
// model. It contains request parsing only; composition remains in the service.
type RuntimeInspectorController struct {
	svc *services.RuntimeInspectorService
}

func NewRuntimeInspectorController(svc *services.RuntimeInspectorService) *RuntimeInspectorController {
	return &RuntimeInspectorController{svc: svc}
}

func (ctrl *RuntimeInspectorController) List(c echo.Context) error {
	tenantID, err := tenantFromHeader(c)
	if err != nil {
		return err
	}
	query := services.RuntimeInstanceQuery{
		Page:     parseInt(c.QueryParam("page"), 1),
		PageSize: parseInt(c.QueryParam("pageSize"), 0),
	}
	if value := c.QueryParam("status"); value != "" {
		status := entities.WorkflowInstanceStatus(value)
		query.Status = &status
	}
	if value := c.QueryParam("workflowId"); value != "" {
		query.WorkflowID = &value
	}
	if value := c.QueryParam("correlationKey"); value != "" {
		query.CorrelationKey = &value
	}
	result, err := ctrl.svc.List(c.Request().Context(), tenantID, query)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"data": result})
}

func (ctrl *RuntimeInspectorController) Detail(c echo.Context) error {
	tenantID, err := tenantFromHeader(c)
	if err != nil {
		return err
	}
	result, err := ctrl.svc.Get(c.Request().Context(), tenantID, c.Param("id"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"data": result})
}

func (ctrl *RuntimeInspectorController) DebugTrace(c echo.Context) error {
	tenantID, err := tenantFromHeader(c)
	if err != nil {
		return err
	}
	result, err := ctrl.svc.DebugTrace(c.Request().Context(), tenantID, c.Param("id"), c.QueryParam("turnId"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"data": result})
}
