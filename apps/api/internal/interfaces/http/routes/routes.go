package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/irosadie/open-state/api/internal/interfaces/http/controllers"
	"github.com/irosadie/open-state/api/internal/interfaces/http/middleware"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	"github.com/irosadie/open-state/api/internal/domain/services"
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

// RegisterCapabilityRoutes registers the tenant-scoped capability registry and
// binding admin endpoints behind auth (PRD §59-64). The tenant is derived from
// the X-Tenant-ID header by the controller, never from the request body.
func RegisterCapabilityRoutes(e *echo.Echo, ctrl *controllers.CapabilityController, repo repositories.IAuthRepository, tokenSvc services.TokenService) {
	g := e.Group("/api/capabilities", middleware.JWT(tokenSvc), middleware.AuthSession(repo, tokenSvc))

	g.GET("", ctrl.List)
	g.POST("", ctrl.Create)
	g.GET("/:id", ctrl.Get)
	g.PATCH("/:id", ctrl.Update)
	g.DELETE("/:id", ctrl.Delete)
	g.GET("/:id/bindings", ctrl.ListBindings)
	g.POST("/:id/bindings", ctrl.CreateBinding)
	g.POST("/:id/test", ctrl.TestInvoke)

	// Binding management is top-level (PRD §60): DELETE /api/bindings/{id}
	b := e.Group("/api/bindings", middleware.JWT(tokenSvc), middleware.AuthSession(repo, tokenSvc))
	b.DELETE("/:id", ctrl.DeleteBinding)
}

func RegisterSystemRoutes(e *echo.Echo, ctrl *controllers.SystemController) {
	e.GET("/health", ctrl.Health)
	e.GET("/", ctrl.AppInfo)
}
