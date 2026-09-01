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

func TestRegisterProjectRoutesAddsListEndpoint(t *testing.T) {
	e := echo.New()
	var authRepo repositories.IAuthRepository
	var tokenService domainsvc.TokenService

	RegisterProjectRoutes(e, controllers.NewProjectController(nil), authRepo, tokenService, nil, (*appservices.AuditWriter)(nil))

	for _, route := range e.Routes() {
		if route.Method == http.MethodGet && route.Path == "/api/projects" {
			return
		}
	}
	t.Fatal("expected GET /api/projects route")
}
