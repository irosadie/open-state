package controllers

import (
	"net/http"

	appservices "github.com/irosadie/open-state/api/internal/application/services"
	"github.com/labstack/echo/v4"
)

// IntentController exposes the read-only canonical intent catalog for Admin
// Console clients. Tenant scope is taken from the existing request header.
type IntentController struct {
	svc *appservices.IntentCatalogService
}

// NewIntentController builds an IntentController.
func NewIntentController(svc *appservices.IntentCatalogService) *IntentController {
	return &IntentController{svc: svc}
}

// List returns published intent mappings for the requested/default project.
func (ctrl *IntentController) List(c echo.Context) error {
	tenantID, err := tenantFromHeader(c)
	if err != nil {
		return err
	}
	result, err := ctrl.svc.List(c.Request().Context(), tenantID, c.QueryParam("projectId"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, result)
}
