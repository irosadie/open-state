package main

import (
	"context"
	"log"

	"github.com/vibecoding-starter/api/internal/application/services"
	usecases "github.com/vibecoding-starter/api/internal/application/use-cases"
	"github.com/vibecoding-starter/api/internal/infrastructure/config"
	infradb "github.com/vibecoding-starter/api/internal/infrastructure/database"
	infrasvc "github.com/vibecoding-starter/api/internal/infrastructure/services"
	"github.com/vibecoding-starter/api/internal/interfaces/http"
	"github.com/vibecoding-starter/api/internal/interfaces/http/controllers"
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
	tokenSvc := infrasvc.NewJwtTokenService(cfg.JWTSecret)
	storageSvc := infrasvc.NewLocalStorageService()
	_ = storageSvc

	// Use cases
	registerUC := usecases.NewRegisterUserUseCase(authRepo, tokenSvc)
	loginUC := usecases.NewLoginUserUseCase(authRepo, tokenSvc)
	logoutUC := usecases.NewLogoutUserUseCase(authRepo, tokenSvc)
	getMeUC := usecases.NewGetCurrentUserUseCase(authRepo)
	healthUC := usecases.NewGetHealthUseCase()
	appInfoUC := usecases.NewGetAppInfoUseCase()

	// Services
	authSvc := services.NewAuthService(registerUC, loginUC, logoutUC, getMeUC)

	// Controllers
	authCtrl := controllers.NewAuthController(authSvc)
	systemCtrl := controllers.NewSystemController(healthUC, appInfoUC)

	// Echo app
	e := http.CreateApp(authCtrl, systemCtrl, authRepo, tokenSvc)

	log.Printf("starting server on :%s", cfg.Port)
	if err := e.Start(":" + cfg.Port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
