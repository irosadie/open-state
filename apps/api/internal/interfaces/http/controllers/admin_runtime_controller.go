package controllers

import (
	"context"
	"net/http"

	"github.com/irosadie/open-state/api/internal/application/dtos"
	appservices "github.com/irosadie/open-state/api/internal/application/services"
	"github.com/irosadie/open-state/api/internal/interfaces/http/middleware"
	"github.com/labstack/echo/v4"
)

// AdminRuntimeController exposes tenant-scoped runtime commands and read-only
// event browsing. Runtime Inspector remains the owner of instance detail UI.
type AdminRuntimeController struct {
	svc *appservices.AdminRuntimeService
}

func NewAdminRuntimeController(svc *appservices.AdminRuntimeService) *AdminRuntimeController {
	return &AdminRuntimeController{svc: svc}
}

func (ctrl *AdminRuntimeController) ListInstances(c echo.Context) error {
	tenantID, err := tenantFromHeader(c)
	if err != nil {
		return err
	}
	result, err := ctrl.svc.ListInstances(c.Request().Context(), tenantID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"data": result})
}

func (ctrl *AdminRuntimeController) Suspend(c echo.Context) error {
	return ctrl.command(c, ctrl.svc.Suspend)
}

func (ctrl *AdminRuntimeController) Resume(c echo.Context) error {
	return ctrl.command(c, ctrl.svc.Resume)
}

func (ctrl *AdminRuntimeController) Retry(c echo.Context) error {
	return ctrl.command(c, ctrl.svc.Retry)
}

func (ctrl *AdminRuntimeController) command(c echo.Context, operation func(context.Context, string, string, string, *string) (*dtos.InstanceDTO, error)) error {
	tenantID, err := tenantFromHeader(c)
	if err != nil {
		return err
	}
	actor, _ := c.Get(middleware.UserIDKey).(string)
	result, err := operation(c.Request().Context(), tenantID, actor, c.Param("id"), correlationID(c))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"data": result})
}

func (ctrl *AdminRuntimeController) ListEvents(c echo.Context) error {
	tenantID, err := tenantFromHeader(c)
	if err != nil {
		return err
	}
	result, err := ctrl.svc.ListEvents(c.Request().Context(), tenantID, appservices.EventBrowserQuery{
		WorkflowInstanceID: c.QueryParam("workflowInstanceId"),
		Type:               c.QueryParam("type"),
		Source:             c.QueryParam("source"),
		CorrelationID:      c.QueryParam("correlationId"),
		Page:               parseInt(c.QueryParam("page"), 1),
		PageSize:           parseInt(c.QueryParam("pageSize"), 0),
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"data": result})
}

func (ctrl *AdminRuntimeController) GetEvent(c echo.Context) error {
	tenantID, err := tenantFromHeader(c)
	if err != nil {
		return err
	}
	result, err := ctrl.svc.GetEvent(c.Request().Context(), tenantID, c.Param("eventId"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"data": result})
}
