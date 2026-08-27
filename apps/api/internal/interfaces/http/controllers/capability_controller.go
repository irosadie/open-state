package controllers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/irosadie/open-state/api/internal/application/dtos"
	appservices "github.com/irosadie/open-state/api/internal/application/services"
	domain "github.com/irosadie/open-state/go-shared/domain"
)

// TenantHeader is the header carrying the tenant id for tenant-scoped routes.
// The tenant is derived from the authenticated request context, never from the
// request body (PRD §74, §96).
const TenantHeader = "X-Tenant-ID"

// CapabilityController exposes the Capability Registry and binding admin
// endpoints. It parses requests, calls the service, and formats responses.
// It contains no business logic (PRD §74).
type CapabilityController struct {
	svc *appservices.CapabilityService
}

// NewCapabilityController builds a CapabilityController.
func NewCapabilityController(svc *appservices.CapabilityService) *CapabilityController {
	return &CapabilityController{svc: svc}
}

func (ctrl *CapabilityController) List(c echo.Context) error {
	tenantID, err := tenantFromHeader(c)
	if err != nil {
		return err
	}
	result, err := ctrl.svc.List(c.Request().Context(), tenantID, c.QueryParam("providerType"), c.QueryParam("status"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": result.Data})
}

func (ctrl *CapabilityController) Get(c echo.Context) error {
	tenantID, err := tenantFromHeader(c)
	if err != nil {
		return err
	}
	result, err := ctrl.svc.FindByID(c.Request().Context(), tenantID, c.Param("id"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": result})
}

func (ctrl *CapabilityController) Create(c echo.Context) error {
	tenantID, err := tenantFromHeader(c)
	if err != nil {
		return err
	}
	var req dtos.CreateCapabilityRequest
	if err := c.Bind(&req); err != nil {
		return domain.NewValidation("invalid request body")
	}
	result, err := ctrl.svc.Create(c.Request().Context(), tenantID, req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, map[string]interface{}{"data": result})
}

func (ctrl *CapabilityController) Update(c echo.Context) error {
	tenantID, err := tenantFromHeader(c)
	if err != nil {
		return err
	}
	var req dtos.UpdateCapabilityRequest
	if err := c.Bind(&req); err != nil {
		return domain.NewValidation("invalid request body")
	}
	result, err := ctrl.svc.Update(c.Request().Context(), tenantID, c.Param("id"), req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": result})
}

func (ctrl *CapabilityController) Delete(c echo.Context) error {
	tenantID, err := tenantFromHeader(c)
	if err != nil {
		return err
	}
	if err := ctrl.svc.Delete(c.Request().Context(), tenantID, c.Param("id")); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "capability disabled"})
}

func (ctrl *CapabilityController) ListBindings(c echo.Context) error {
	tenantID, err := tenantFromHeader(c)
	if err != nil {
		return err
	}
	result, err := ctrl.svc.ListBindings(c.Request().Context(), tenantID, c.Param("id"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": result})
}

func (ctrl *CapabilityController) CreateBinding(c echo.Context) error {
	tenantID, err := tenantFromHeader(c)
	if err != nil {
		return err
	}
	var req dtos.CreateBindingRequest
	if err := c.Bind(&req); err != nil {
		return domain.NewValidation("invalid request body")
	}
	result, err := ctrl.svc.Bind(c.Request().Context(), tenantID, c.Param("id"), req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, map[string]interface{}{"data": result})
}

func (ctrl *CapabilityController) DeleteBinding(c echo.Context) error {
	tenantID, err := tenantFromHeader(c)
	if err != nil {
		return err
	}
	if err := ctrl.svc.Unbind(c.Request().Context(), tenantID, c.Param("id")); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "binding removed"})
}

func (ctrl *CapabilityController) TestInvoke(c echo.Context) error {
	tenantID, err := tenantFromHeader(c)
	if err != nil {
		return err
	}
	var req dtos.TestInvocationRequest
	if err := c.Bind(&req); err != nil {
		return domain.NewValidation("invalid request body")
	}
	result, err := ctrl.svc.TestInvoke(c.Request().Context(), tenantID, c.Param("id"), req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": result})
}

func tenantFromHeader(c echo.Context) (string, error) {
	tenantID := c.Request().Header.Get(TenantHeader)
	if tenantID == "" {
		return "", domain.NewUnauthorized("missing tenant header")
	}
	return tenantID, nil
}
