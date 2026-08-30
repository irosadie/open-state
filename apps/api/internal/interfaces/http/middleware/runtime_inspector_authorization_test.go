package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	appservices "github.com/irosadie/open-state/api/internal/application/services"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/labstack/echo/v4"
)

type runtimeInspectorRoleRepo struct {
	role entities.UserRole
}

func (r runtimeInspectorRoleRepo) FindRoleByUserAndTenant(context.Context, string, string) (entities.UserRole, error) {
	return r.role, nil
}

func TestRuntimeInspectorPermissionsAreIndependent(t *testing.T) {
	tests := []struct {
		name          string
		role          entities.UserRole
		required      appservices.Permission
		wantForbidden bool
	}{
		{name: "viewer can inspect instance", role: entities.UserRoleViewer, required: "instance:read"},
		{name: "viewer cannot inspect debug trace", role: entities.UserRoleViewer, required: "debug:read", wantForbidden: true},
		{name: "operator can inspect debug trace", role: entities.UserRoleOperator, required: "debug:read"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/api/runtime/instances/instance-1", nil)
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)
			ctx.Set(UserIDKey, "user-1")
			ctx.Request().Header.Set(TenantHeader, "tenant-1")

			authz := appservices.NewAuthorizationService(runtimeInspectorRoleRepo{role: tt.role})
			called := false
			handler := RequirePermission(authz, tt.required, nil)(func(echo.Context) error {
				called = true
				return nil
			})
			err := handler(ctx)

			if tt.wantForbidden {
				if err == nil {
					t.Fatal("expected permission denial")
				}
				if called {
					t.Fatal("expected denied request not to reach handler")
				}
				return
			}
			if err != nil {
				t.Fatalf("expected permission to pass, got %v", err)
			}
			if !called {
				t.Fatal("expected authorized request to reach handler")
			}
		})
	}
}
