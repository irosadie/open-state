package routes

import (
	"net/http"
	"testing"

	appservices "github.com/irosadie/open-state/api/internal/application/services"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	domainsvc "github.com/irosadie/open-state/api/internal/domain/services"
	"github.com/irosadie/open-state/api/internal/interfaces/http/controllers"
	"github.com/labstack/echo/v4"
)

func TestRegisterIntentRoutesAddsReadEndpoint(t *testing.T) {
	e := echo.New()
	var authRepo repositories.IAuthRepository
	var tokenService domainsvc.TokenService

	RegisterIntentRoutes(e, controllers.NewIntentController(nil), authRepo, tokenService, nil, (*appservices.AuditWriter)(nil))

	for _, route := range e.Routes() {
		if route.Method == http.MethodGet && route.Path == "/api/intents" {
			return
		}
	}
	t.Fatal("expected GET /api/intents route")
}
