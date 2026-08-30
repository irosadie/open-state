package http

import (
	"log/slog"

	appservices "github.com/irosadie/open-state/api/internal/application/services"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	domainsvc "github.com/irosadie/open-state/api/internal/domain/services"
	"github.com/irosadie/open-state/api/internal/interfaces/http/controllers"
	"github.com/irosadie/open-state/api/internal/interfaces/http/middleware"
	"github.com/irosadie/open-state/api/internal/interfaces/http/routes"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
)

func CreateApp(
	authCtrl *controllers.AuthController,
	systemCtrl *controllers.SystemController,
	capabilityCtrl *controllers.CapabilityController,
	workflowCtrl *controllers.WorkflowController,
	auditCtrl *controllers.AuditController,
	ssoCtrl *controllers.SSOController,
	repo repositories.IAuthRepository,
	tokenSvc domainsvc.TokenService,
	authz *appservices.AuthorizationService,
	audit *appservices.AuditWriter,
	loginLimiter domainsvc.RateLimiter,
	registerLimiter domainsvc.RateLimiter,
	logger *slog.Logger,
	metricsRec middleware.MetricsRecorder,
	adminIdentityCtrl *controllers.AdminIdentityController,
	adminRuntimeCtrl *controllers.AdminRuntimeController,
	runtimeCtrls ...*controllers.RuntimeInspectorController,
) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HTTPErrorHandler = middleware.ErrorHandler

	// Global middleware
	e.Use(echomw.Recover())
	e.Use(echomw.RequestID())
	e.Use(middleware.RequestLogger(logger))
	e.Use(middleware.Metrics(metricsRec))

	// Routes
	routes.RegisterSystemRoutes(e, systemCtrl)
	routes.RegisterAuthRoutes(e, authCtrl, repo, tokenSvc, loginLimiter, registerLimiter)
	routes.RegisterSSORoutes(e, ssoCtrl)
	routes.RegisterCapabilityRoutes(e, capabilityCtrl, repo, tokenSvc, authz, audit)
	routes.RegisterWorkflowRoutes(e, workflowCtrl, repo, tokenSvc, authz, audit)
	routes.RegisterAuditRoutes(e, auditCtrl, repo, tokenSvc, authz, audit)
	if adminIdentityCtrl != nil && adminRuntimeCtrl != nil {
		routes.RegisterAdminRoutes(e, adminIdentityCtrl, adminRuntimeCtrl, repo, tokenSvc, authz, audit)
	}
	if len(runtimeCtrls) > 0 && runtimeCtrls[0] != nil {
		routes.RegisterRuntimeInspectorRoutes(e, runtimeCtrls[0], repo, tokenSvc, authz, audit)
	}

	return e
}
