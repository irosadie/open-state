package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/vibecoding-starter/api/internal/interfaces/http/controllers"
	"github.com/vibecoding-starter/api/internal/interfaces/http/middleware"
	"github.com/vibecoding-starter/api/internal/domain/repositories"
	"github.com/vibecoding-starter/api/internal/domain/services"
)

func RegisterAuthRoutes(e *echo.Echo, ctrl *controllers.AuthController, repo repositories.IAuthRepository, tokenSvc services.TokenService) {
	auth := e.Group("/api/auth")

	// Public routes
	auth.POST("/register", ctrl.Register)
	auth.POST("/login", ctrl.Login)

	// Protected routes
	protected := auth.Group("", middleware.JWT(tokenSvc), middleware.AuthSession(repo, tokenSvc))
	protected.POST("/logout", ctrl.Logout)
	protected.GET("/me", ctrl.Me)
}

func RegisterSystemRoutes(e *echo.Echo, ctrl *controllers.SystemController) {
	e.GET("/health", ctrl.Health)
	e.GET("/", ctrl.AppInfo)
}
