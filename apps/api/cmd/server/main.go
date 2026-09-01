package main

import (
	"context"
	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/irosadie/open-state/api/internal/application/services"
	usecases "github.com/irosadie/open-state/api/internal/application/use-cases"
	domainsvc "github.com/irosadie/open-state/api/internal/domain/services"
	infracap "github.com/irosadie/open-state/api/internal/infrastructure/capability"
	"github.com/irosadie/open-state/api/internal/infrastructure/config"
	infradb "github.com/irosadie/open-state/api/internal/infrastructure/database"
	infralog "github.com/irosadie/open-state/api/internal/infrastructure/logging"
	inframetrics "github.com/irosadie/open-state/api/internal/infrastructure/metrics"
	infroidc "github.com/irosadie/open-state/api/internal/infrastructure/oidc"
	infrl "github.com/irosadie/open-state/api/internal/infrastructure/ratelimit"
	infrasvc "github.com/irosadie/open-state/api/internal/infrastructure/services"
	infratrace "github.com/irosadie/open-state/api/internal/infrastructure/tracing"
	"github.com/irosadie/open-state/api/internal/interfaces/http"
	"github.com/irosadie/open-state/api/internal/interfaces/http/controllers"
	httpmw "github.com/irosadie/open-state/api/internal/interfaces/http/middleware"
)

func main() {
	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config error", "error", err.Error())
		return
	}

	logger := infralog.New(infralog.Config{Format: cfg.LogFormat, Level: cfg.LogLevel})
	slog.SetDefault(logger)

	// OpenTelemetry tracing (PRD §84). No-op when no OTLP endpoint is set.
	traceShutdown := infratrace.Setup(ctx, infratrace.Config{
		OTLPEndpoint: cfg.OTel.OTLPEndpoint,
		ServiceName:  cfg.OTel.ServiceName,
	}, logger)
	defer func() { _ = traceShutdown(context.Background()) }()

	// Prometheus metrics (PRD §84): RED + runtime + application metrics.
	var metricsRegistry *inframetrics.Registry
	if cfg.MetricsEnabled {
		metricsRegistry = inframetrics.New()
	}
	var metricsRecorder httpmw.MetricsRecorder
	var auditMetrics services.AuditMetrics
	if metricsRegistry != nil {
		metricsRecorder = metricsRegistry
		auditMetrics = metricsRegistry
	}

	pool, err := config.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("database error", "error", err.Error())
		return
	}
	defer pool.Close()

	// Infrastructure
	authRepo := infradb.NewPgxAuthRepository(pool)
	// Composed persistence port (ADR-001): the single pgx-backed adapter exposing
	// all six repository interfaces (workflow, instance, event, context,
	// capability, audit). Application/engine services depend on these interfaces;
	// the adapter is the portability seam. Constructed here to establish the
	// dependency wiring and consumed by subsequent slices.
	adapter := infradb.NewPostgresAdapter(pool)
	tokenSvc := infrasvc.NewJwtTokenService(cfg.JWTSecret)
	storageSvc := infrasvc.NewLocalStorageService()
	_ = storageSvc
	_ = adapter

	// RBAC authorization service (PRD 80, 81): resolves a user's tenant role from
	// role_assignments and checks permissions against the domain matrix.
	authzSvc := services.NewAuthorizationService(adapter.Roles())

	// Audit writer (PRD 50): append-only, tenant-isolated audit trail for
	// important operations and authorization denials.
	auditWriter := services.NewAuditWriter(adapter.Audit(), logger, auditMetrics)

	// Audit query service + controller (PRD 50): filtered, paginated audit trail.
	auditSvc := services.NewAuditService(adapter.Audit())
	runtimeInspectorSvc := services.NewRuntimeInspectorService(
		adapter.Instances(), adapter.RuntimeRead(), adapter.Events(), adapter.Context(), adapter.Audit(), adapter.RuntimeTraces(),
	)
	orchestratorSvc := services.NewOrchestratorService(adapter.Instances(), adapter.Events(), adapter.Context(), adapter.Capabilities(), adapter.Workflows())
	adminIdentitySvc := services.NewAdminIdentityService(adapter.Admin(), auditWriter)
	adminRuntimeSvc := services.NewAdminRuntimeService(orchestratorSvc, adapter.EventBrowser(), auditWriter)

	// Rate limiters (PRD 83): token-bucket per operation, in-memory. Login and
	// register protect the public auth endpoints from brute-force/mass-registration;
	// capability limits protect provider invocation from abuse.
	loginLimiter := infrl.NewTokenBucket(infrl.Config{Rate: cfg.RateLimit.Login.Rate, Burst: cfg.RateLimit.Login.Burst})
	registerLimiter := infrl.NewTokenBucket(infrl.Config{Rate: cfg.RateLimit.Register.Rate, Burst: cfg.RateLimit.Register.Burst})
	capabilityLimiter := infrl.NewTokenBucket(infrl.Config{Rate: cfg.RateLimit.Capability.Rate, Burst: cfg.RateLimit.Capability.Burst})

	// Use cases
	registerUC := usecases.NewRegisterUserUseCase(authRepo, adapter.Roles(), tokenSvc)
	loginUC := usecases.NewLoginUserUseCase(authRepo, tokenSvc)
	logoutUC := usecases.NewLogoutUserUseCase(authRepo, tokenSvc)
	getMeUC := usecases.NewGetCurrentUserUseCase(authRepo)
	healthUC := usecases.NewGetHealthUseCase()
	appInfoUC := usecases.NewGetAppInfoUseCase()

	// Capability admin service (tenant-scoped registry + bindings + sandbox test)
	capSvc := services.NewCapabilityService(adapter.Capabilities(), infracap.MockProviderResolver{}, infracap.JSONSchemaValidator{}, auditWriter, capabilityLimiter)

	// Builder API service (workflow-definition drafts + publish + versions, PRD 146)
	builderSvc := services.NewBuilderService(adapter.Workflows(), adapter.Projects(), auditWriter)
	intentSvc := services.NewIntentService(adapter.Intents(), adapter.Workflows())
	intentCatalogSvc := services.NewIntentCatalogService(intentSvc, adapter.Projects())
	projectSvc := services.NewProjectService(adapter.Projects())
	simulationSvc := services.NewSimulationService()
	apiKeySvc := services.NewAPIKeyService(adapter.APIKeys(), adapter.Projects(), auditWriter, cfg.MCPAPIKeyPepper)

	// Services
	authSvc := services.NewAuthService(registerUC, loginUC, logoutUC, getMeUC, authzSvc)

	// SSO providers (PRD §79): build OIDC adapters for enabled providers only.
	ssoProviders := make(map[string]domainsvc.OIDCProvider)
	for name, pc := range map[string]config.OIDCProviderConfig{
		"google": cfg.SSO.Google,
		"github": cfg.SSO.GitHub,
		"entra":  cfg.SSO.Entra,
	} {
		if !pc.Enabled() {
			continue
		}
		provider, err := infroidc.New(ctx, infroidc.ProviderConfig{
			Name:         name,
			ClientID:     pc.ClientID,
			ClientSecret: pc.ClientSecret,
			IssuerURL:    pc.IssuerURL,
			RedirectURI:  pc.RedirectURI,
			Scopes:       pc.Scopes,
		})
		if err != nil {
			logger.Warn("sso provider disabled", "provider", name, "error", err.Error())
			continue
		}
		ssoProviders[name] = provider
	}
	ssoSvc := services.NewSSOService(adapter.Identities(), authRepo, adapter.Roles(), tokenSvc, ssoProviders, auditWriter)
	ssoCtrl := controllers.NewSSOController(ssoSvc, cfg.SSO.BaseURL)

	// Controllers
	authCtrl := controllers.NewAuthController(authSvc)
	systemCtrl := controllers.NewSystemController(healthUC, appInfoUC)
	capCtrl := controllers.NewCapabilityController(capSvc)
	workflowCtrl := controllers.NewWorkflowController(builderSvc, simulationSvc)
	intentCtrl := controllers.NewIntentController(intentCatalogSvc)
	projectCtrl := controllers.NewProjectController(projectSvc)
	auditCtrl := controllers.NewAuditController(auditSvc)
	adminIdentityCtrl := controllers.NewAdminIdentityController(adminIdentitySvc)
	adminRuntimeCtrl := controllers.NewAdminRuntimeController(adminRuntimeSvc)
	runtimeInspectorCtrl := controllers.NewRuntimeInspectorController(runtimeInspectorSvc)
	apiKeyCtrl := controllers.NewAPIKeyController(apiKeySvc)

	// Echo app
	e := http.CreateApp(authCtrl, systemCtrl, capCtrl, workflowCtrl, intentCtrl, auditCtrl, ssoCtrl, authRepo, tokenSvc, authzSvc, auditWriter, loginLimiter, registerLimiter, logger, metricsRecorder, adminIdentityCtrl, adminRuntimeCtrl, apiKeyCtrl, projectCtrl, runtimeInspectorCtrl)

	// Distributed tracing middleware (PRD §84): server span per request + traceparent
	// extraction. Applied after CreateApp to keep the interfaces layer free of
	// infrastructure imports.
	e.Use(infratrace.NewHTTPTrace(logger).Middleware())

	// Hardened CORS + security headers (PRD §74, §139). Applied after CreateApp.
	e.Use(httpmw.CORS(httpmw.CORSConfig{AllowedOrigins: cfg.Security.AllowedOrigins}))
	e.Use(httpmw.SecurityHeaders(httpmw.SecurityHeadersConfig{
		ContentSecurityPolicy: cfg.Security.CSP,
		EnableHSTS:            cfg.Security.EnableHSTS,
		HSTSMaxAge:            cfg.Security.HSTSMaxAge,
	}))

	// Prometheus metrics endpoint (PRD §84).
	if metricsRegistry != nil {
		e.GET("/metrics", echo.WrapHandler(promhttp.HandlerFor(metricsRegistry.Prometheus(), promhttp.HandlerOpts{})))
		logger.Info("metrics endpoint enabled", "path", "/metrics")
	}

	logger.Info("starting server", "port", cfg.Port)
	if err := e.Start(":" + cfg.Port); err != nil {
		logger.Error("server error", "error", err.Error())
	}
}
