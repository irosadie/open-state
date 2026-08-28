package routes

import (
	appservices "github.com/irosadie/open-state/api/internal/application/services"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	domainsvc "github.com/irosadie/open-state/api/internal/domain/services"
	"github.com/irosadie/open-state/api/internal/interfaces/http/controllers"
	"github.com/irosadie/open-state/api/internal/interfaces/http/middleware"
	"github.com/labstack/echo/v4"
)

func RegisterAuthRoutes(e *echo.Echo, ctrl *controllers.AuthController, repo repositories.IAuthRepository, tokenSvc domainsvc.TokenService, loginLimiter domainsvc.RateLimiter, registerLimiter domainsvc.RateLimiter) {
	auth := e.Group("/api/auth")

	// Public routes — rate-limited to protect against brute-force and
	// mass-registration abuse (PRD §83).
	auth.POST("/register", ctrl.Register, middleware.RateLimit(registerLimiter, middleware.RegisterKey))
	auth.POST("/login", ctrl.Login, middleware.RateLimit(loginLimiter, middleware.LoginKey))

	// Protected routes
	protected := auth.Group("", middleware.JWT(tokenSvc), middleware.AuthSession(repo, tokenSvc))
	protected.POST("/logout", ctrl.Logout)
	protected.GET("/me", ctrl.Me)
}

// RegisterCapabilityRoutes registers the tenant-scoped capability registry and
// binding admin endpoints behind auth + RBAC (PRD 59-64, 80-81). The tenant is
// derived from the X-Tenant-ID header by the middleware/controller, never from
// the request body.
func RegisterCapabilityRoutes(e *echo.Echo, ctrl *controllers.CapabilityController, repo repositories.IAuthRepository, tokenSvc domainsvc.TokenService, authz *appservices.AuthorizationService, audit *appservices.AuditWriter) {
	auth := []echo.MiddlewareFunc{middleware.JWT(tokenSvc), middleware.AuthSession(repo, tokenSvc)}
	g := e.Group("/api/capabilities", auth...)

	g.GET("", ctrl.List, middleware.RequirePermission(authz, "capability:read", audit))
	g.POST("", ctrl.Create, middleware.RequirePermission(authz, "capability:create", audit))
	g.GET("/:id", ctrl.Get, middleware.RequirePermission(authz, "capability:read", audit))
	g.PATCH("/:id", ctrl.Update, middleware.RequirePermission(authz, "capability:update", audit))
	g.DELETE("/:id", ctrl.Delete, middleware.RequirePermission(authz, "capability:delete", audit))
	g.GET("/:id/bindings", ctrl.ListBindings, middleware.RequirePermission(authz, "binding:read", audit))
	g.POST("/:id/bindings", ctrl.CreateBinding, middleware.RequirePermission(authz, "binding:create", audit))
	g.POST("/:id/test", ctrl.TestInvoke, middleware.RequirePermission(authz, "capability:invoke", audit))

	// Binding management is top-level (PRD 60): DELETE /api/bindings/{id}
	b := e.Group("/api/bindings", auth...)
	b.DELETE("/:id", ctrl.DeleteBinding, middleware.RequirePermission(authz, "binding:delete", audit))
}

// RegisterWorkflowRoutes registers the tenant+project-scoped Builder API
// (PRD 146) behind auth + RBAC. The tenant comes from the X-Tenant-ID header;
// the project is optional and defaults to the tenant's default project.
func RegisterWorkflowRoutes(e *echo.Echo, ctrl *controllers.WorkflowController, repo repositories.IAuthRepository, tokenSvc domainsvc.TokenService, authz *appservices.AuthorizationService, audit *appservices.AuditWriter) {
	auth := []echo.MiddlewareFunc{middleware.JWT(tokenSvc), middleware.AuthSession(repo, tokenSvc)}
	g := e.Group("/api/workflows", auth...)

	g.GET("", ctrl.List, middleware.RequirePermission(authz, "workflow:read", audit))
	g.POST("", ctrl.Create, middleware.RequirePermission(authz, "workflow:create", audit))
	g.GET("/:id", ctrl.Get, middleware.RequirePermission(authz, "workflow:read", audit))
	g.PATCH("/:id", ctrl.Update, middleware.RequirePermission(authz, "workflow:update", audit))
	g.POST("/:id/publish", ctrl.Publish, middleware.RequirePermission(authz, "workflow:publish", audit))
	g.GET("/:id/versions", ctrl.ListVersions, middleware.RequirePermission(authz, "workflow:read", audit))
}

// RegisterAuditRoutes registers the tenant-scoped audit trail query API
// (PRD 50) behind auth + the audit:read RBAC guard.
func RegisterAuditRoutes(e *echo.Echo, ctrl *controllers.AuditController, repo repositories.IAuthRepository, tokenSvc domainsvc.TokenService, authz *appservices.AuthorizationService, audit *appservices.AuditWriter) {
	g := e.Group("/api/audit", middleware.JWT(tokenSvc), middleware.AuthSession(repo, tokenSvc))
	g.GET("", ctrl.List, middleware.RequirePermission(authz, "audit:read", audit))
}

func RegisterSystemRoutes(e *echo.Echo, ctrl *controllers.SystemController) {
	e.GET("/health", ctrl.Health)
	e.GET("/", ctrl.AppInfo)
}
