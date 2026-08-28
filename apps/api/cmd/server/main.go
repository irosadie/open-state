package main

import (
	"context"
	"log"

	"github.com/irosadie/open-state/api/internal/application/services"
	usecases "github.com/irosadie/open-state/api/internal/application/use-cases"
	infracap "github.com/irosadie/open-state/api/internal/infrastructure/capability"
	"github.com/irosadie/open-state/api/internal/infrastructure/config"
	infradb "github.com/irosadie/open-state/api/internal/infrastructure/database"
	infrl "github.com/irosadie/open-state/api/internal/infrastructure/ratelimit"
	infrasvc "github.com/irosadie/open-state/api/internal/infrastructure/services"
	"github.com/irosadie/open-state/api/internal/interfaces/http"
	"github.com/irosadie/open-state/api/internal/interfaces/http/controllers"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	ctx := context.Background()
	pool, err := config.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database error: %v", err)
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
	auditWriter := services.NewAuditWriter(adapter.Audit())

	// Audit query service + controller (PRD 50): filtered, paginated audit trail.
	auditSvc := services.NewAuditService(adapter.Audit())

	// Rate limiters (PRD 83): token-bucket per operation, in-memory. Login and
	// register protect the public auth endpoints from brute-force/mass-registration;
	// capability limits protect provider invocation from abuse.
	loginLimiter := infrl.NewTokenBucket(infrl.Config{Rate: cfg.RateLimit.Login.Rate, Burst: cfg.RateLimit.Login.Burst})
	registerLimiter := infrl.NewTokenBucket(infrl.Config{Rate: cfg.RateLimit.Register.Rate, Burst: cfg.RateLimit.Register.Burst})
	capabilityLimiter := infrl.NewTokenBucket(infrl.Config{Rate: cfg.RateLimit.Capability.Rate, Burst: cfg.RateLimit.Capability.Burst})

	// Use cases
	registerUC := usecases.NewRegisterUserUseCase(authRepo, tokenSvc)
	loginUC := usecases.NewLoginUserUseCase(authRepo, tokenSvc)
	logoutUC := usecases.NewLogoutUserUseCase(authRepo, tokenSvc)
	getMeUC := usecases.NewGetCurrentUserUseCase(authRepo)
	healthUC := usecases.NewGetHealthUseCase()
	appInfoUC := usecases.NewGetAppInfoUseCase()

	// Capability admin service (tenant-scoped registry + bindings + sandbox test)
	capSvc := services.NewCapabilityService(adapter.Capabilities(), infracap.MockProviderResolver{}, infracap.JSONSchemaValidator{}, auditWriter, capabilityLimiter)

	// Builder API service (workflow-definition drafts + publish + versions, PRD 146)
	builderSvc := services.NewBuilderService(adapter.Workflows(), adapter.Projects(), auditWriter)

	// Services
	authSvc := services.NewAuthService(registerUC, loginUC, logoutUC, getMeUC, authzSvc)

	// Controllers
	authCtrl := controllers.NewAuthController(authSvc)
	systemCtrl := controllers.NewSystemController(healthUC, appInfoUC)
	capCtrl := controllers.NewCapabilityController(capSvc)
	workflowCtrl := controllers.NewWorkflowController(builderSvc)
	auditCtrl := controllers.NewAuditController(auditSvc)

	// Echo app
	e := http.CreateApp(authCtrl, systemCtrl, capCtrl, workflowCtrl, auditCtrl, authRepo, tokenSvc, authzSvc, auditWriter, loginLimiter, registerLimiter)

	log.Printf("starting server on :%s", cfg.Port)
	if err := e.Start(":" + cfg.Port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
