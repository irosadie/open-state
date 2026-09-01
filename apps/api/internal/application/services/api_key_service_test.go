package services

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/irosadie/open-state/api/internal/application/dtos"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	domain "github.com/irosadie/open-state/go-shared/domain"
)

type memoryAPIKeyRepository struct {
	keys map[string]entities.APIKey
}

func (r *memoryAPIKeyRepository) Create(_ context.Context, input repositories.APIKeyCreateInput) (*entities.APIKey, error) {
	if r.keys == nil {
		r.keys = make(map[string]entities.APIKey)
	}
	key := entities.APIKey{
		ID: "key-1", TenantID: input.TenantID, Name: input.Name, Prefix: input.Prefix,
		KeyVerifier: append([]byte(nil), input.KeyVerifier...), ProjectIDs: append([]string(nil), input.ProjectIDs...),
		DefaultProjectID: input.DefaultProjectID, Scopes: append([]entities.MCPAPIScope(nil), input.Scopes...),
		ExpiresAt: input.ExpiresAt, CreatedBy: input.CreatedBy, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	r.keys[key.Prefix] = key
	return &key, nil
}

func (r *memoryAPIKeyRepository) FindByPrefix(_ context.Context, prefix string) (*entities.APIKey, error) {
	key, ok := r.keys[prefix]
	if !ok {
		return nil, domain.NewNotFound("API key")
	}
	copy := cloneAPIKey(key)
	return &copy, nil
}

func (r *memoryAPIKeyRepository) ListByTenant(_ context.Context, tenantID string) ([]entities.APIKey, error) {
	keys := make([]entities.APIKey, 0, len(r.keys))
	for _, key := range r.keys {
		if key.TenantID == tenantID {
			keys = append(keys, cloneAPIKey(key))
		}
	}
	return keys, nil
}

func (r *memoryAPIKeyRepository) Revoke(_ context.Context, tenantID, keyID string) (*entities.APIKey, error) {
	for prefix, key := range r.keys {
		if key.TenantID == tenantID && key.ID == keyID && key.RevokedAt == nil {
			now := time.Now()
			key.RevokedAt = &now
			r.keys[prefix] = key
			copy := cloneAPIKey(key)
			return &copy, nil
		}
	}
	return nil, domain.NewNotFound("API key")
}

func (r *memoryAPIKeyRepository) TouchLastUsed(_ context.Context, keyID string) error {
	for prefix, key := range r.keys {
		if key.ID == keyID {
			now := time.Now()
			key.LastUsedAt = &now
			r.keys[prefix] = key
		}
	}
	return nil
}

type memoryProjectRepository struct {
	projects map[string]entities.Project
}

func (r memoryProjectRepository) Create(context.Context, string, string, string, entities.ProjectStatus) (*entities.Project, error) {
	return nil, domain.NewInternal("not implemented")
}
func (r memoryProjectRepository) FindByID(_ context.Context, tenantID, id string) (*entities.Project, error) {
	project, ok := r.projects[id]
	if !ok || project.TenantID != tenantID {
		return nil, domain.NewNotFound("project")
	}
	return &project, nil
}
func (r memoryProjectRepository) FindBySlug(context.Context, string, string) (*entities.Project, error) {
	return nil, domain.NewNotFound("project")
}
func (r memoryProjectRepository) ListByTenant(context.Context, string) ([]entities.Project, error) {
	return nil, nil
}

func TestAPIKeyServiceCreatesOneTimeSecretAndAuthenticates(t *testing.T) {
	repo := &memoryAPIKeyRepository{}
	svc := NewAPIKeyService(repo, memoryProjectRepository{projects: map[string]entities.Project{
		"project-1": {ID: "project-1", TenantID: "tenant-1"},
	}}, nil, "test-pepper-must-have-at-least-thirty-two-characters")

	created, err := svc.Create(context.Background(), "tenant-1", "user-1", dtos.CreateAPIKeyRequest{
		Name: "Claude desktop", ProjectIDs: []string{"project-1"}, DefaultProjectID: apiKeyStringPtr("project-1"),
		Scopes: []string{"state:read"},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.Key == "" || created.APIKey.Prefix == "" {
		t.Fatal("expected raw key and safe prefix")
	}
	if jsonBytes, err := json.Marshal(created.APIKey); err != nil || string(jsonBytes) == created.Key || strings.Contains(string(jsonBytes), created.Key) {
		t.Fatalf("safe metadata leaked raw key: %s", jsonBytes)
	}
	stored := repo.keys[created.APIKey.Prefix]
	if string(stored.KeyVerifier) == created.Key {
		t.Fatal("raw API key was persisted instead of a verifier")
	}

	principal, err := svc.Authenticate(context.Background(), created.Key)
	if err != nil {
		t.Fatalf("authenticate: %v", err)
	}
	if principal.TenantID != "tenant-1" || !principal.AllowsProject("project-1") || !principal.HasScope(entities.MCPAPIScopeStateRead) {
		t.Fatalf("unexpected principal: %+v", principal)
	}
}

func TestAPIKeyServiceRejectsInvalidScopeAndRevokedKey(t *testing.T) {
	repo := &memoryAPIKeyRepository{}
	svc := NewAPIKeyService(repo, memoryProjectRepository{projects: map[string]entities.Project{
		"project-1": {ID: "project-1", TenantID: "tenant-1"},
	}}, nil, "test-pepper-must-have-at-least-thirty-two-characters")
	_, err := svc.Create(context.Background(), "tenant-1", "user-1", dtos.CreateAPIKeyRequest{
		Name: "invalid", ProjectIDs: []string{"project-1"}, Scopes: []string{"tenant:admin"},
	})
	if err == nil {
		t.Fatal("expected invalid scope to be rejected")
	}

	created, err := svc.Create(context.Background(), "tenant-1", "user-1", dtos.CreateAPIKeyRequest{
		Name: "valid", ProjectIDs: []string{"project-1"}, Scopes: []string{"state:read"},
	})
	if err != nil {
		t.Fatalf("create valid key: %v", err)
	}
	if _, err := svc.Revoke(context.Background(), "tenant-1", "user-1", created.APIKey.ID); err != nil {
		t.Fatalf("revoke: %v", err)
	}
	if _, err := svc.Authenticate(context.Background(), created.Key); err == nil {
		t.Fatal("expected revoked API key authentication to fail")
	}
}

func cloneAPIKey(key entities.APIKey) entities.APIKey {
	key.KeyVerifier = append([]byte(nil), key.KeyVerifier...)
	key.ProjectIDs = append([]string(nil), key.ProjectIDs...)
	key.Scopes = append([]entities.MCPAPIScope(nil), key.Scopes...)
	return key
}

func apiKeyStringPtr(value string) *string { return &value }
