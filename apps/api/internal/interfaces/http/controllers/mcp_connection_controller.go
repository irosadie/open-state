package controllers

import (
	"net/http"

	"github.com/irosadie/open-state/api/internal/application/dtos"
	appservices "github.com/irosadie/open-state/api/internal/application/services"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/interfaces/http/middleware"
	domain "github.com/irosadie/open-state/go-shared/domain"
	"github.com/labstack/echo/v4"
)

// MCPConnectionController exposes the project MCP connection registry.
type MCPConnectionController struct {
	svc   *appservices.MCPConnectionService
	oauth *appservices.MCPOAuthService
}

func NewMCPConnectionController(svc *appservices.MCPConnectionService, oauth ...*appservices.MCPOAuthService) *MCPConnectionController {
	var oauthSvc *appservices.MCPOAuthService
	if len(oauth) > 0 {
		oauthSvc = oauth[0]
	}
	return &MCPConnectionController{svc: svc, oauth: oauthSvc}
}

func (ctrl *MCPConnectionController) List(c echo.Context) error {
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

func (ctrl *MCPConnectionController) Get(c echo.Context) error {
	tenantID, projectID, err := mcpScope(c)
	if err != nil {
		return err
	}
	result, err := ctrl.svc.Get(c.Request().Context(), tenantID, projectID, c.Param("id"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"data": result})
}

func (ctrl *MCPConnectionController) Create(c echo.Context) error {
	tenantID, projectID, err := mcpScope(c)
	if err != nil {
		return err
	}
	var req dtos.CreateMCPConnectionRequest
	if err := c.Bind(&req); err != nil {
		return domain.NewValidation("invalid request body")
	}
	actor, _ := c.Get(middleware.UserIDKey).(string)
	result, err := ctrl.svc.Create(c.Request().Context(), tenantID, projectID, actor, req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, map[string]any{"data": result})
}

func (ctrl *MCPConnectionController) Update(c echo.Context) error {
	tenantID, projectID, err := mcpScope(c)
	if err != nil {
		return err
	}
	var req dtos.UpdateMCPConnectionRequest
	if err := c.Bind(&req); err != nil {
		return domain.NewValidation("invalid request body")
	}
	actor, _ := c.Get(middleware.UserIDKey).(string)
	result, err := ctrl.svc.Update(c.Request().Context(), tenantID, projectID, c.Param("id"), actor, req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"data": result})
}

func (ctrl *MCPConnectionController) Delete(c echo.Context) error {
	tenantID, projectID, err := mcpScope(c)
	if err != nil {
		return err
	}
	actor, _ := c.Get(middleware.UserIDKey).(string)
	if err := ctrl.svc.Delete(c.Request().Context(), tenantID, projectID, c.Param("id"), actor); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "MCP connection deleted"})
}

func (ctrl *MCPConnectionController) Enable(c echo.Context) error {
	return ctrl.setStatus(c, entities.MCPConnectionEnabled)
}
func (ctrl *MCPConnectionController) Disable(c echo.Context) error {
	return ctrl.setStatus(c, entities.MCPConnectionDisabled)
}

func (ctrl *MCPConnectionController) setStatus(c echo.Context, status entities.MCPConnectionStatus) error {
	tenantID, projectID, err := mcpScope(c)
	if err != nil {
		return err
	}
	actor, _ := c.Get(middleware.UserIDKey).(string)
	result, err := ctrl.svc.SetStatus(c.Request().Context(), tenantID, projectID, c.Param("id"), actor, status)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"data": result})
}

func (ctrl *MCPConnectionController) Test(c echo.Context) error {
	tenantID, projectID, err := mcpScope(c)
	if err != nil {
		return err
	}
	actor, _ := c.Get(middleware.UserIDKey).(string)
	result, err := ctrl.svc.Test(c.Request().Context(), tenantID, projectID, c.Param("id"), actor)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"data": result})
}

func (ctrl *MCPConnectionController) Diagnose(c echo.Context) error {
	tenantID, projectID, err := mcpScope(c)
	if err != nil {
		return err
	}
	actor, _ := c.Get(middleware.UserIDKey).(string)
	result, err := ctrl.svc.Diagnose(c.Request().Context(), tenantID, projectID, c.Param("id"), actor)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"data": result})
}

func (ctrl *MCPConnectionController) ResetHealth(c echo.Context) error {
	tenantID, projectID, err := mcpScope(c)
	if err != nil {
		return err
	}
	actor, _ := c.Get(middleware.UserIDKey).(string)
	result, err := ctrl.svc.ResetHealth(c.Request().Context(), tenantID, projectID, c.Param("id"), actor)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"data": result})
}

func (ctrl *MCPConnectionController) CredentialRotate(c echo.Context) error {
	tenantID, projectID, err := mcpScope(c)
	if err != nil {
		return err
	}
	var req struct {
		CredentialValue string `json:"credentialValue"`
	}
	if err := c.Bind(&req); err != nil {
		return domain.NewValidation("invalid request body")
	}
	actor, _ := c.Get(middleware.UserIDKey).(string)
	result, err := ctrl.svc.RotateCredential(c.Request().Context(), tenantID, projectID, c.Param("id"), actor, req.CredentialValue)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"data": result})
}

func (ctrl *MCPConnectionController) CredentialRevoke(c echo.Context) error {
	tenantID, projectID, err := mcpScope(c)
	if err != nil {
		return err
	}
	actor, _ := c.Get(middleware.UserIDKey).(string)
	result, err := ctrl.svc.RevokeCredential(c.Request().Context(), tenantID, projectID, c.Param("id"), actor)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"data": result})
}

func (ctrl *MCPConnectionController) CredentialStatus(c echo.Context) error {
	tenantID, projectID, err := mcpScope(c)
	if err != nil {
		return err
	}
	result, err := ctrl.svc.CredentialStatus(c.Request().Context(), tenantID, projectID, c.Param("id"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"data": result})
}

func (ctrl *MCPConnectionController) OAuthStart(c echo.Context) error {
	if ctrl.oauth == nil {
		return domain.NewInternal("OAuth lifecycle is not configured")
	}
	tenantID, projectID, err := mcpScope(c)
	if err != nil {
		return err
	}
	actor, _ := c.Get(middleware.UserIDKey).(string)
	result, err := ctrl.oauth.Start(c.Request().Context(), tenantID, projectID, c.Param("id"), actor)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"data": result})
}

func (ctrl *MCPConnectionController) OAuthCallback(c echo.Context) error {
	if ctrl.oauth == nil {
		return domain.NewInternal("OAuth lifecycle is not configured")
	}
	tenantID, projectID, err := mcpScope(c)
	if err != nil {
		return err
	}
	actor, _ := c.Get(middleware.UserIDKey).(string)
	result, err := ctrl.oauth.Callback(c.Request().Context(), tenantID, projectID, c.Param("id"), actor, c.QueryParam("state"), c.QueryParam("code"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"data": result})
}

func (ctrl *MCPConnectionController) OAuthDisconnect(c echo.Context) error {
	if ctrl.oauth == nil {
		return domain.NewInternal("OAuth lifecycle is not configured")
	}
	tenantID, projectID, err := mcpScope(c)
	if err != nil {
		return err
	}
	actor, _ := c.Get(middleware.UserIDKey).(string)
	result, err := ctrl.oauth.Disconnect(c.Request().Context(), tenantID, projectID, c.Param("id"), actor)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"data": result})
}

func (ctrl *MCPConnectionController) OAuthStatus(c echo.Context) error {
	if ctrl.oauth == nil {
		return domain.NewInternal("OAuth lifecycle is not configured")
	}
	tenantID, projectID, err := mcpScope(c)
	if err != nil {
		return err
	}
	result, err := ctrl.oauth.Status(c.Request().Context(), tenantID, projectID, c.Param("id"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"data": result})
}

func mcpScope(c echo.Context) (string, string, error) {
	tenantID, err := tenantFromHeader(c)
	if err != nil {
		return "", "", err
	}
	projectID := c.Param("projectId")
	if projectID == "" {
		return "", "", domain.NewValidation("projectId is required")
	}
	return tenantID, projectID, nil
}
