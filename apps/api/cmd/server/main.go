package main

import (
	"context"
	"log"

	"github.com/irosadie/open-state/api/internal/application/services"
	usecases "github.com/irosadie/open-state/api/internal/application/use-cases"
	infracap "github.com/irosadie/open-state/api/internal/infrastructure/capability"
	"github.com/irosadie/open-state/api/internal/infrastructure/config"
	infradb "github.com/irosadie/open-state/api/internal/infrastructure/database"
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

	// Use cases
	registerUC := usecases.NewRegisterUserUseCase(authRepo, tokenSvc)
	loginUC := usecases.NewLoginUserUseCase(authRepo, tokenSvc)
	logoutUC := usecases.NewLogoutUserUseCase(authRepo, tokenSvc)
	getMeUC := usecases.NewGetCurrentUserUseCase(authRepo)
	healthUC := usecases.NewGetHealthUseCase()
	appInfoUC := usecases.NewGetAppInfoUseCase()

	// Capability admin service (tenant-scoped registry + bindings + sandbox test)
	capSvc := services.NewCapabilityService(adapter.Capabilities(), infracap.MockProviderResolver{}, infracap.JSONSchemaValidator{})

	// Builder API service (workflow-definition drafts + publish + versions, PRD 146)
	builderSvc := services.NewBuilderService(adapter.Workflows(), adapter.Projects())

	// Services
	authSvc := services.NewAuthService(registerUC, loginUC, logoutUC, getMeUC)

	// Controllers
	authCtrl := controllers.NewAuthController(authSvc)
	systemCtrl := controllers.NewSystemController(healthUC, appInfoUC)
	capCtrl := controllers.NewCapabilityController(capSvc)
	workflowCtrl := controllers.NewWorkflowController(builderSvc)

	// Echo app
	e := http.CreateApp(authCtrl, systemCtrl, capCtrl, workflowCtrl, authRepo, tokenSvc)

	log.Printf("starting server on :%s", cfg.Port)
	if err := e.Start(":" + cfg.Port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
