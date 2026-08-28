package controllers

import (
	"net/http"
	"strings"

	"github.com/irosadie/open-state/api/internal/application/dtos"
	appservices "github.com/irosadie/open-state/api/internal/application/services"
	usecases "github.com/irosadie/open-state/api/internal/application/use-cases"
	"github.com/irosadie/open-state/api/internal/interfaces/http/middleware"
	domain "github.com/irosadie/open-state/go-shared/domain"
	"github.com/labstack/echo/v4"
)

type AuthController struct {
	authSvc *appservices.AuthService
}

func NewAuthController(authSvc *appservices.AuthService) *AuthController {
	return &AuthController{authSvc: authSvc}
}

func (ctrl *AuthController) Register(c echo.Context) error {
	var req dtos.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	user, err := ctrl.authSvc.Register(c.Request().Context(), usecases.RegisterUserInput{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{"data": user})
}

func (ctrl *AuthController) Login(c echo.Context) error {
	var req dtos.LoginRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	result, err := ctrl.authSvc.Login(c.Request().Context(), usecases.LoginUserInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"data": result})
}

func (ctrl *AuthController) Logout(c echo.Context) error {
	header := c.Request().Header.Get("Authorization")
	token := strings.TrimPrefix(header, "Bearer ")

	if err := ctrl.authSvc.Logout(c.Request().Context(), token); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "logged out"})
}

func (ctrl *AuthController) Me(c echo.Context) error {
	userID, _ := c.Get(middleware.UserIDKey).(string)
	tenantID := c.Request().Header.Get(middleware.TenantHeader)
	if tenantID == "" {
		return domain.NewUnauthorized("missing tenant header")
	}

	user, err := ctrl.authSvc.GetCurrentUserForTenant(c.Request().Context(), userID, tenantID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"data": user})
}
