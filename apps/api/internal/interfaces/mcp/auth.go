package mcp

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	appservices "github.com/irosadie/open-state/api/internal/application/services"
	"github.com/irosadie/open-state/api/internal/domain/entities"
)

type principalContextKey struct{}

// WithAPIKeyPrincipal stores an authenticated machine principal in a request
// context. It is exported for HTTP integration tests as well as the middleware.
func WithAPIKeyPrincipal(ctx context.Context, principal entities.APIKeyPrincipal) context.Context {
	return context.WithValue(ctx, principalContextKey{}, principal)
}

func principalFromContext(ctx context.Context) (entities.APIKeyPrincipal, bool) {
	principal, ok := ctx.Value(principalContextKey{}).(entities.APIKeyPrincipal)
	return principal, ok
}

// APIKeyAuthentication authenticates State MCP requests before the protocol
// transport receives them. It never returns credential-specific failure detail.
func APIKeyAuthentication(auth *appservices.APIKeyService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if auth == nil {
				writeUnauthorized(w)
				return
			}
			rawKey, ok := bearerToken(r.Header.Get("Authorization"))
			if !ok {
				writeUnauthorized(w)
				return
			}
			principal, err := auth.Authenticate(r.Context(), rawKey)
			if err != nil {
				writeUnauthorized(w)
				return
			}
			next.ServeHTTP(w, r.WithContext(WithAPIKeyPrincipal(r.Context(), *principal)))
		})
	}
}

type authorizedToolHandler func(context.Context, entities.APIKeyPrincipal, mcp.CallToolRequest) (*mcp.CallToolResult, error)

func authorizedTool(deps Dependencies, scope entities.MCPAPIScope, next authorizedToolHandler) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		principal, ok := principalFromContext(ctx)
		if !ok {
			return toolError(errors.New("unauthenticated MCP request"))
		}
		if !principal.HasScope(scope) {
			recordDenied(ctx, deps, principal, "missing_scope")
			return toolError(errors.New("permission denied"))
		}
		return next(ctx, principal, req)
	}
}

func projectForPrincipal(ctx context.Context, deps Dependencies, principal entities.APIKeyPrincipal, requested string) (string, error) {
	projectID := strings.TrimSpace(requested)
	if projectID == "" {
		if principal.DefaultProjectID == nil {
			recordDenied(ctx, deps, principal, "missing_default_project")
			return "", errors.New("project is required for this API key")
		}
		projectID = *principal.DefaultProjectID
	}
	if !principal.AllowsProject(projectID) {
		recordDenied(ctx, deps, principal, "project_not_allowed")
		return "", errors.New("project is not allowed")
	}
	return projectID, nil
}

func recordDenied(ctx context.Context, deps Dependencies, principal entities.APIKeyPrincipal, reason string) {
	if deps.APIKeyAuth != nil {
		deps.APIKeyAuth.RecordDeniedToolAccess(ctx, principal, reason)
	}
}

func bearerToken(header string) (string, bool) {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", false
	}
	return parts[1], true
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
}
