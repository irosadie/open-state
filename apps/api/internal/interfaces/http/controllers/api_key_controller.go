package controllers

import (
	"net/http"

	"github.com/irosadie/open-state/api/internal/application/dtos"
	appservices "github.com/irosadie/open-state/api/internal/application/services"
	"github.com/irosadie/open-state/api/internal/interfaces/http/middleware"
	domain "github.com/irosadie/open-state/go-shared/domain"
	"github.com/labstack/echo/v4"
)

// APIKeyController exposes tenant administrators' State MCP API-key lifecycle.
type APIKeyController struct {
	svc *appservices.APIKeyService
}

// NewAPIKeyController builds an APIKeyController.
func NewAPIKeyController(svc *appservices.APIKeyService) *APIKeyController {
	return &APIKeyController{svc: svc}
}

func (ctrl *APIKeyController) List(c echo.Context) error {
	tenantID, err := tenantFromHeader(c)
	if err != nil {
		return err
	}
	result, err := ctrl.svc.List(c.Request().Context(), tenantID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": result})
}

func (ctrl *APIKeyController) Create(c echo.Context) error {
	tenantID, err := tenantFromHeader(c)
	if err != nil {
		return err
	}
	actor, _ := c.Get(middleware.UserIDKey).(string)
	var req dtos.CreateAPIKeyRequest
	if err := c.Bind(&req); err != nil {
		return domain.NewValidation("invalid request body")
	}
	result, err := ctrl.svc.Create(c.Request().Context(), tenantID, actor, req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, map[string]interface{}{"data": result})
}

func (ctrl *APIKeyController) Revoke(c echo.Context) error {
	tenantID, err := tenantFromHeader(c)
	if err != nil {
		return err
	}
	actor, _ := c.Get(middleware.UserIDKey).(string)
	result, err := ctrl.svc.Revoke(c.Request().Context(), tenantID, actor, c.Param("id"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": result})
}
