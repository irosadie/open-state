package mcp

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	appservices "github.com/irosadie/open-state/api/internal/application/services"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

func TestAPIKeyAuthenticationRejectsMissingBearer(t *testing.T) {
	nextCalled := false
	handler := APIKeyAuthentication(nil)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		nextCalled = true
	}))
	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusUnauthorized || nextCalled {
		t.Fatalf("expected unauthorized request without protocol execution, status=%d next=%v", res.Code, nextCalled)
	}
}

func TestAuthorizedToolRejectsMissingScope(t *testing.T) {
	deps := testDeps()
	principal := testPrincipal("tenant-1")
	principal.Scopes = []entities.MCPAPIScope{entities.MCPAPIScopeStateRead}
	server := newAuthenticatedTestServer(NewServer(deps), principal)
	defer server.Close()

	client, err := client.NewStreamableHttpClient(server.URL + "/mcp")
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	if _, err := client.Initialize(context.Background(), mcp.InitializeRequest{}); err != nil {
		t.Fatalf("initialize: %v", err)
	}
	result, err := client.CallTool(context.Background(), mcp.CallToolRequest{Params: mcp.CallToolParams{
		Name: "start_workflow", Arguments: map[string]any{"workflow": "wf-1"},
	}})
	if err != nil {
		t.Fatalf("call tool: %v", err)
	}
	content := result.Content[0].(mcp.TextContent)
	if !strings.Contains(content.Text, "permission denied") {
		t.Fatalf("expected scope denial, got %+v", result)
	}
}

func TestAPIKeyAuthenticationAcceptsActiveAndRejectsRevokedKey(t *testing.T) {
	rawKey := "osk_test_machine-secret"
	pepper := "test-pepper-must-have-at-least-thirty-two-characters"
	mac := hmac.New(sha256.New, []byte(pepper))
	_, _ = mac.Write([]byte(rawKey))
	repo := &staticAPIKeyRepository{key: entities.APIKey{
		ID: "key-1", TenantID: "tenant-1", Prefix: "osk_test", KeyVerifier: mac.Sum(nil),
		ProjectIDs: []string{"project-1"}, Scopes: []entities.MCPAPIScope{entities.MCPAPIScopeStateRead},
	}}
	auth := appservices.NewAPIKeyService(repo, nil, nil, pepper)
	principalTenant := ""
	handler := APIKeyAuthentication(auth)(http.HandlerFunc(func(_ http.ResponseWriter, request *http.Request) {
		principal, ok := principalFromContext(request.Context())
		if ok {
			principalTenant = principal.TenantID
		}
	}))

	request := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	request.Header.Set("Authorization", "Bearer "+rawKey)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK || principalTenant != "tenant-1" {
		t.Fatalf("expected active key to reach MCP with tenant principal, status=%d tenant=%q", response.Code, principalTenant)
	}

	now := time.Now()
	repo.key.RevokedAt = &now
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected revoked key to be rejected, got %d", response.Code)
	}
}

func TestProjectScopeRejectsAndTenantArgumentIsIgnored(t *testing.T) {
	resolver := &capturingIntentResolver{}
	deps := testDeps()
	deps.IntentResolver = resolver
	principal := testPrincipal("tenant-1")
	principal.ProjectIDs = []string{"project-padel"}
	server := newAuthenticatedTestServer(NewServer(deps), principal)
	defer server.Close()

	client, err := client.NewStreamableHttpClient(server.URL + "/mcp")
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	if _, err := client.Initialize(context.Background(), mcp.InitializeRequest{}); err != nil {
		t.Fatalf("initialize: %v", err)
	}
	denied, err := client.CallTool(context.Background(), mcp.CallToolRequest{Params: mcp.CallToolParams{
		Name: "list_intents", Arguments: map[string]any{"project": "project-food"},
	}})
	if err != nil {
		t.Fatalf("call denied project: %v", err)
	}
	if !strings.Contains(denied.Content[0].(mcp.TextContent).Text, "project is not allowed") {
		t.Fatalf("expected project denial, got %+v", denied)
	}

	_, err = client.CallTool(context.Background(), mcp.CallToolRequest{Params: mcp.CallToolParams{
		Name: "list_intents", Arguments: map[string]any{"tenant": "tenant-other", "project": "project-padel"},
	}})
	if err != nil {
		t.Fatalf("call tenant impersonation attempt: %v", err)
	}
	if resolver.tenantID != "tenant-1" {
		t.Fatalf("tool trusted tenant argument instead of principal: %q", resolver.tenantID)
	}
}

type staticAPIKeyRepository struct{ key entities.APIKey }

func (r *staticAPIKeyRepository) Create(context.Context, repositories.APIKeyCreateInput) (*entities.APIKey, error) {
	return nil, errors.New("not implemented")
}
func (r *staticAPIKeyRepository) FindByPrefix(_ context.Context, prefix string) (*entities.APIKey, error) {
	if prefix != r.key.Prefix {
		return nil, errors.New("not found")
	}
	copy := r.key
	return &copy, nil
}
func (r *staticAPIKeyRepository) ListByTenant(context.Context, string) ([]entities.APIKey, error) {
	return nil, errors.New("not implemented")
}
func (r *staticAPIKeyRepository) Revoke(context.Context, string, string) (*entities.APIKey, error) {
	return nil, errors.New("not implemented")
}
func (r *staticAPIKeyRepository) TouchLastUsed(context.Context, string) error { return nil }

type capturingIntentResolver struct{ tenantID string }

func (r *capturingIntentResolver) ListIntents(_ context.Context, tenantID, _ string) ([]entities.Intent, error) {
	r.tenantID = tenantID
	return nil, nil
}
func (r *capturingIntentResolver) ResolveIntent(context.Context, string, string, string) (*entities.Workflow, error) {
	return nil, errors.New("not implemented")
}
