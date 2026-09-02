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
	g.POST("/simulate", ctrl.Simulate, middleware.RequirePermission(authz, "workflow:simulate", audit))
	g.GET("/:id", ctrl.Get, middleware.RequirePermission(authz, "workflow:read", audit))
	g.PATCH("/:id", ctrl.Update, middleware.RequirePermission(authz, "workflow:update", audit))
	g.POST("/:id/publish", ctrl.Publish, middleware.RequirePermission(authz, "workflow:publish", audit))
	g.GET("/:id/versions", ctrl.ListVersions, middleware.RequirePermission(authz, "workflow:read", audit))
	g.GET("/:id/versions/compare", ctrl.CompareVersions, middleware.RequirePermission(authz, "workflow:read", audit))
	g.GET("/:id/versions/:versionNo", ctrl.GetVersion, middleware.RequirePermission(authz, "workflow:read", audit))
}

// RegisterIntentRoutes registers the read-only canonical intent catalog behind
// the same workflow read permission used by the workflow inventory.
func RegisterIntentRoutes(e *echo.Echo, ctrl *controllers.IntentController, repo repositories.IAuthRepository, tokenSvc domainsvc.TokenService, authz *appservices.AuthorizationService, audit *appservices.AuditWriter) {
	g := e.Group("/api/intents", middleware.JWT(tokenSvc), middleware.AuthSession(repo, tokenSvc))
	g.GET("", ctrl.List, middleware.RequirePermission(authz, "workflow:read", audit))
}

// RegisterProjectRoutes exposes read-only tenant project discovery for the
// Admin Console flow and State MCP API-key creation.
func RegisterProjectRoutes(e *echo.Echo, ctrl *controllers.ProjectController, repo repositories.IAuthRepository, tokenSvc domainsvc.TokenService, authz *appservices.AuthorizationService, audit *appservices.AuditWriter) {
	g := e.Group("/api/projects", middleware.JWT(tokenSvc), middleware.AuthSession(repo, tokenSvc))
	g.GET("", ctrl.List, middleware.RequirePermission(authz, "workflow:read", audit))
}

// RegisterMCPConnectionRoutes exposes project-owned external MCP connections.
func RegisterMCPConnectionRoutes(e *echo.Echo, ctrl *controllers.MCPConnectionController, repo repositories.IAuthRepository, tokenSvc domainsvc.TokenService, authz *appservices.AuthorizationService, audit *appservices.AuditWriter) {
	g := e.Group("/api/projects/:projectId/mcp-connections", middleware.JWT(tokenSvc), middleware.AuthSession(repo, tokenSvc))
	g.GET("", ctrl.List, middleware.RequirePermission(authz, "mcp_connection:read", audit))
	g.POST("", ctrl.Create, middleware.RequirePermission(authz, "mcp_connection:create", audit))
	g.GET("/:id", ctrl.Get, middleware.RequirePermission(authz, "mcp_connection:read", audit))
	g.PATCH("/:id", ctrl.Update, middleware.RequirePermission(authz, "mcp_connection:update", audit))
	g.DELETE("/:id", ctrl.Delete, middleware.RequirePermission(authz, "mcp_connection:delete", audit))
	g.POST("/:id/enable", ctrl.Enable, middleware.RequirePermission(authz, "mcp_connection:enable", audit))
	g.POST("/:id/disable", ctrl.Disable, middleware.RequirePermission(authz, "mcp_connection:disable", audit))
	g.POST("/:id/test", ctrl.Test, middleware.RequirePermission(authz, "mcp_connection:test", audit))
	g.POST("/:id/diagnose", ctrl.Diagnose, middleware.RequirePermission(authz, "mcp_connection:test", audit))
	g.POST("/:id/reset-health", ctrl.ResetHealth, middleware.RequirePermission(authz, "mcp_connection:update", audit))
	g.POST("/:id/credentials/rotate", ctrl.CredentialRotate, middleware.RequirePermission(authz, "mcp_connection:update", audit))
	g.POST("/:id/credentials/revoke", ctrl.CredentialRevoke, middleware.RequirePermission(authz, "mcp_connection:update", audit))
	g.GET("/:id/credentials/status", ctrl.CredentialStatus, middleware.RequirePermission(authz, "mcp_connection:read", audit))
	g.POST("/:id/oauth/start", ctrl.OAuthStart, middleware.RequirePermission(authz, "mcp_connection:update", audit))
	g.GET("/:id/oauth/callback", ctrl.OAuthCallback, middleware.RequirePermission(authz, "mcp_connection:update", audit))
	g.POST("/:id/oauth/disconnect", ctrl.OAuthDisconnect, middleware.RequirePermission(authz, "mcp_connection:update", audit))
	g.GET("/:id/oauth/status", ctrl.OAuthStatus, middleware.RequirePermission(authz, "mcp_connection:read", audit))
}

// RegisterMCPToolCatalogRoutes exposes explicit tools/list discovery and
// project-scoped tool enablement. There is intentionally no tools/call route.
func RegisterMCPToolCatalogRoutes(e *echo.Echo, ctrl *controllers.MCPToolCatalogController, repo repositories.IAuthRepository, tokenSvc domainsvc.TokenService, authz *appservices.AuthorizationService, audit *appservices.AuditWriter) {
	g := e.Group("/api/projects/:projectId/mcp-connections/:id/tools", middleware.JWT(tokenSvc), middleware.AuthSession(repo, tokenSvc))
	g.GET("", ctrl.List, middleware.RequirePermission(authz, "mcp_connection:read", audit))
	g.POST("/refresh", ctrl.Refresh, middleware.RequirePermission(authz, "mcp_connection:discover", audit))
	g.PATCH("/:toolName", ctrl.SetEnabled, middleware.RequirePermission(authz, "mcp_connection:tool:update", audit))
}

// RegisterProjectCapabilityMCPBindingRoutes exposes safe project catalog
// options and explicit logical capability bindings for State Builder.
func RegisterProjectCapabilityMCPBindingRoutes(e *echo.Echo, ctrl *controllers.ProjectCapabilityMCPBindingController, repo repositories.IAuthRepository, tokenSvc domainsvc.TokenService, authz *appservices.AuthorizationService, audit *appservices.AuditWriter) {
	g := e.Group("/api/projects/:projectId", middleware.JWT(tokenSvc), middleware.AuthSession(repo, tokenSvc))
	g.GET("/mcp-tool-options", ctrl.ListOptions, middleware.RequirePermission(authz, "mcp_connection:read", audit))
	g.GET("/mcp-capability-bindings", ctrl.List, middleware.RequirePermission(authz, "mcp_connection:read", audit))
	g.PUT("/capabilities/:capabilityId/mcp-binding", ctrl.Upsert, middleware.RequirePermission(authz, "workflow:update", audit))
	g.DELETE("/capabilities/:capabilityId/mcp-binding", ctrl.Delete, middleware.RequirePermission(authz, "workflow:update", audit))
}

// RegisterAPIKeyRoutes exposes the tenant-admin lifecycle for State MCP
// machine credentials. Human JWT/RBAC remains the authority for these routes.
func RegisterAPIKeyRoutes(e *echo.Echo, ctrl *controllers.APIKeyController, repo repositories.IAuthRepository, tokenSvc domainsvc.TokenService, authz *appservices.AuthorizationService, audit *appservices.AuditWriter) {
	g := e.Group("/api/api-keys", middleware.JWT(tokenSvc), middleware.AuthSession(repo, tokenSvc))
	g.GET("", ctrl.List, middleware.RequirePermission(authz, "api_key:read", audit))
	g.POST("", ctrl.Create, middleware.RequirePermission(authz, "api_key:create", audit))
	g.POST("/:id/revoke", ctrl.Revoke, middleware.RequirePermission(authz, "api_key:revoke", audit))
}

// RegisterAuditRoutes registers the tenant-scoped audit trail query API
// (PRD 50) behind auth + the audit:read RBAC guard.
func RegisterAuditRoutes(e *echo.Echo, ctrl *controllers.AuditController, repo repositories.IAuthRepository, tokenSvc domainsvc.TokenService, authz *appservices.AuthorizationService, audit *appservices.AuditWriter) {
	g := e.Group("/api/audit", middleware.JWT(tokenSvc), middleware.AuthSession(repo, tokenSvc))
	g.GET("", ctrl.List, middleware.RequirePermission(authz, "audit:read", audit))
}

// RegisterRuntimeInspectorRoutes registers the read-only, tenant-scoped Runtime
// Inspector and separately protected Debug View routes.
func RegisterRuntimeInspectorRoutes(e *echo.Echo, ctrl *controllers.RuntimeInspectorController, repo repositories.IAuthRepository, tokenSvc domainsvc.TokenService, authz *appservices.AuthorizationService, audit *appservices.AuditWriter) {
	g := e.Group("/api/runtime/instances", middleware.JWT(tokenSvc), middleware.AuthSession(repo, tokenSvc))
	g.GET("", ctrl.List, middleware.RequirePermission(authz, "instance:read", audit))
	g.GET("/:id", ctrl.Detail, middleware.RequirePermission(authz, "instance:read", audit))
	g.GET("/:id/debug-trace", ctrl.DebugTrace, middleware.RequirePermission(authz, "debug:read", audit))
}

// RegisterAdminRoutes registers the permission-aware Admin Console backend.
// Every route derives tenant scope from X-Tenant-ID and uses the exact existing
// RBAC permission for the operation.
func RegisterAdminRoutes(e *echo.Echo, identity *controllers.AdminIdentityController, runtime *controllers.AdminRuntimeController, repo repositories.IAuthRepository, tokenSvc domainsvc.TokenService, authz *appservices.AuthorizationService, audit *appservices.AuditWriter) {
	auth := []echo.MiddlewareFunc{middleware.JWT(tokenSvc), middleware.AuthSession(repo, tokenSvc)}

	identityRoutes := e.Group("/api/admin", auth...)
	identityRoutes.GET("/tenant", identity.GetTenant, middleware.RequirePermission(authz, "tenant:read", audit))
	identityRoutes.PATCH("/tenant", identity.UpdateTenant, middleware.RequirePermission(authz, "tenant:update", audit))
	identityRoutes.GET("/members", identity.ListMembers, middleware.RequirePermission(authz, "user:read", audit))
	identityRoutes.PUT("/members/:userId/role", identity.UpdateMemberRole, middleware.RequirePermission(authz, "user:update", audit))
	identityRoutes.DELETE("/members/:userId", identity.RemoveMember, middleware.RequirePermission(authz, "user:delete", audit))

	runtimeRoutes := e.Group("/api/admin", auth...)
	runtimeRoutes.GET("/instances", runtime.ListInstances, middleware.RequirePermission(authz, "instance:read", audit))
	runtimeRoutes.POST("/instances/:id/suspend", runtime.Suspend, middleware.RequirePermission(authz, "instance:suspend", audit))
	runtimeRoutes.POST("/instances/:id/resume", runtime.Resume, middleware.RequirePermission(authz, "instance:resume", audit))
	runtimeRoutes.POST("/instances/:id/retry", runtime.Retry, middleware.RequirePermission(authz, "instance:retry", audit))
	runtimeRoutes.GET("/events", runtime.ListEvents, middleware.RequirePermission(authz, "instance:read", audit))
	runtimeRoutes.GET("/events/:eventId", runtime.GetEvent, middleware.RequirePermission(authz, "instance:read", audit))
}

// RegisterSSORoutes registers the external OIDC SSO login endpoints (PRD §79).
func RegisterSSORoutes(e *echo.Echo, ctrl *controllers.SSOController) {
	auth := e.Group("/api/auth/sso")
	auth.GET("/providers", ctrl.Providers)
	auth.GET("/:provider", ctrl.Start)
	auth.GET("/:provider/callback", ctrl.Callback)
}

func RegisterSystemRoutes(e *echo.Echo, ctrl *controllers.SystemController) {
	e.GET("/health", ctrl.Health)
	e.GET("/", ctrl.AppInfo)
}
