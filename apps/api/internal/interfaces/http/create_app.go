package http

import (
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"github.com/irosadie/open-state/api/internal/interfaces/http/controllers"
	"github.com/irosadie/open-state/api/internal/interfaces/http/middleware"
	"github.com/irosadie/open-state/api/internal/interfaces/http/routes"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	"github.com/irosadie/open-state/api/internal/domain/services"
)

func CreateApp(
	authCtrl *controllers.AuthController,
	systemCtrl *controllers.SystemController,
	capabilityCtrl *controllers.CapabilityController,
	repo repositories.IAuthRepository,
	tokenSvc services.TokenService,
) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HTTPErrorHandler = middleware.ErrorHandler

	// Global middleware
	e.Use(echomw.Logger())
	e.Use(echomw.Recover())
	e.Use(echomw.CORS())
	e.Use(echomw.RequestID())

	// Routes
	routes.RegisterSystemRoutes(e, systemCtrl)
	routes.RegisterAuthRoutes(e, authCtrl, repo, tokenSvc)
	routes.RegisterCapabilityRoutes(e, capabilityCtrl, repo, tokenSvc)

	return e
}
