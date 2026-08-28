package middleware

import (
	"context"
	"net/http"

	appservices "github.com/irosadie/open-state/api/internal/application/services"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	domain "github.com/irosadie/open-state/go-shared/domain"
	"github.com/labstack/echo/v4"
)

// TenantHeader is the header carrying the tenant id for tenant-scoped routes.
// It is derived from the authenticated request context, never from the request
// body (PRD 74, 96).
const TenantHeader = "X-Tenant-ID"

// RequirePermission is an Echo middleware that authorizes the authenticated
// user for the request tenant against a required permission, using the
// tenant-scoped role-permission matrix (PRD 80, 81). It MUST run after the JWT
// (and AuthSession) middleware so the user id is present in the context.
//
// Semantics (distinct 401 vs 403):
//   - unauthenticated / missing user id  -> 401
//   - missing tenant header              -> 401
//   - authenticated but permission denied -> 403
//
// When a denial is recorded (audit != nil), an authorization.denied audit entry
// is appended (PRD 50, 80).
func RequirePermission(authz *appservices.AuthorizationService, required appservices.Permission, audit *appservices.AuditWriter) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userID, _ := c.Get(UserIDKey).(string)
			if userID == "" {
				return &echo.HTTPError{Code: http.StatusUnauthorized, Message: "unauthorized"}
			}

			tenantID := c.Request().Header.Get(TenantHeader)
			if tenantID == "" {
				return domain.NewUnauthorized("missing tenant header")
			}

			if err := authz.Require(c.Request().Context(), userID, tenantID, required); err != nil {
				// Audit authorization denials (PRD 50, 80). Never record secrets.
				if audit != nil {
					audit.Write(c.Request().Context(), tenantID, userID, entities.AuditActionAuthDenied,
						resourceTypeForPermission(required), string(required), nil, nil, nil)
				}
				return err // FORBIDDEN domain error -> 403 via ErrorHandler
			}
			return next(c)
		}
	}
}

// resourceTypeForPermission derives a coarse resource type from a permission
// string (e.g. "workflow:create" -> "workflow") for audit recording.
func resourceTypeForPermission(p appservices.Permission) string {
	for i := 0; i < len(p); i++ {
		if p[i] == ':' {
			return string(p[:i])
		}
	}
	return string(p)
}

// auditRoleChange appends an RBAC role-mutation audit entry (PRD 50, 80, 81).
// It is a helper for callers that mutate role_assignments.
func auditRoleChange(ctx context.Context, audit *appservices.AuditWriter, tenantID, actor, targetUserID string, action entities.AuditAction, before, after any) {
	if audit == nil {
		return
	}
	audit.Write(ctx, tenantID, actor, action, "role_assignment", targetUserID, before, after, nil)
}
