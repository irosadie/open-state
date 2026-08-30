package controllers

import (
	"net/http"

	"github.com/irosadie/open-state/api/internal/application/dtos"
	appservices "github.com/irosadie/open-state/api/internal/application/services"
	"github.com/irosadie/open-state/api/internal/interfaces/http/middleware"
	"github.com/labstack/echo/v4"
)

// AdminIdentityController exposes current-tenant profile and membership
// operations. Authorization is attached by the route layer.
type AdminIdentityController struct {
	svc *appservices.AdminIdentityService
}

func NewAdminIdentityController(svc *appservices.AdminIdentityService) *AdminIdentityController {
	return &AdminIdentityController{svc: svc}
}

func (ctrl *AdminIdentityController) GetTenant(c echo.Context) error {
	tenantID, err := tenantFromHeader(c)
	if err != nil {
		return err
	}
	result, err := ctrl.svc.GetTenant(c.Request().Context(), tenantID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"data": result})
}

func (ctrl *AdminIdentityController) UpdateTenant(c echo.Context) error {
	tenantID, err := tenantFromHeader(c)
	if err != nil {
		return err
	}
	var req dtos.UpdateTenantRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "invalid request body")
	}
	actor, _ := c.Get(middleware.UserIDKey).(string)
	result, err := ctrl.svc.UpdateTenant(c.Request().Context(), tenantID, actor, req, correlationID(c))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"data": result})
}

func (ctrl *AdminIdentityController) ListMembers(c echo.Context) error {
	tenantID, err := tenantFromHeader(c)
	if err != nil {
		return err
	}
	result, err := ctrl.svc.ListMemberships(
		c.Request().Context(), tenantID, c.QueryParam("search"),
		parseInt(c.QueryParam("page"), 1), parseInt(c.QueryParam("pageSize"), 0),
	)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"data": result})
}

func (ctrl *AdminIdentityController) UpdateMemberRole(c echo.Context) error {
	tenantID, err := tenantFromHeader(c)
	if err != nil {
		return err
	}
	var req dtos.UpdateMembershipRoleRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "invalid request body")
	}
	actor, _ := c.Get(middleware.UserIDKey).(string)
	result, err := ctrl.svc.UpdateMembershipRole(c.Request().Context(), tenantID, actor, c.Param("userId"), req.Role, correlationID(c))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"data": result})
}

func (ctrl *AdminIdentityController) RemoveMember(c echo.Context) error {
	tenantID, err := tenantFromHeader(c)
	if err != nil {
		return err
	}
	actor, _ := c.Get(middleware.UserIDKey).(string)
	if err := ctrl.svc.RemoveMembership(c.Request().Context(), tenantID, actor, c.Param("userId"), correlationID(c)); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "membership removed"})
}

func correlationID(c echo.Context) *string {
	value := c.Request().Header.Get("X-Correlation-ID")
	if value == "" {
		return nil
	}
	return &value
}
