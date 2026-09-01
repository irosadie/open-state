package services

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/irosadie/open-state/api/internal/application/dtos"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	domain "github.com/irosadie/open-state/go-shared/domain"
)

const apiKeyPrefixMarker = "osk"

// APIKeyService manages State MCP machine credentials and resolves their
// authenticated principals. The raw secret exists only in Create's response.
type APIKeyService struct {
	repo     repositories.IAPIKeyRepository
	projects repositories.IProjectRepository
	audit    *AuditWriter
	pepper   []byte
	now      func() time.Time
}

// NewAPIKeyService constructs the API-key lifecycle service. The caller must
// provide a non-empty server-side pepper from configuration.
func NewAPIKeyService(repo repositories.IAPIKeyRepository, projects repositories.IProjectRepository, audit *AuditWriter, pepper string) *APIKeyService {
	return &APIKeyService{repo: repo, projects: projects, audit: audit, pepper: []byte(pepper), now: time.Now}
}

// Create issues one machine credential for a tenant and returns its raw secret
// exactly once together with safe metadata.
func (s *APIKeyService) Create(ctx context.Context, tenantID, actor string, req dtos.CreateAPIKeyRequest) (*dtos.CreateAPIKeyResponse, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, domain.NewValidation("API key name is required")
	}
	projectIDs, err := uniqueNonEmpty(req.ProjectIDs)
	if err != nil {
		return nil, err
	}
	if len(projectIDs) == 0 {
		return nil, domain.NewValidation("at least one project is required")
	}
	for _, projectID := range projectIDs {
		if _, err := s.projects.FindByID(ctx, tenantID, projectID); err != nil {
			return nil, domain.NewValidation("project is not available to this tenant")
		}
	}
	if req.DefaultProjectID != nil && !contains(projectIDs, *req.DefaultProjectID) {
		return nil, domain.NewValidation("default project must be in projectIds")
	}
	scopes, err := parseScopes(req.Scopes)
	if err != nil {
		return nil, err
	}
	if req.ExpiresAt != nil && !req.ExpiresAt.After(s.now()) {
		return nil, domain.NewValidation("expiresAt must be in the future")
	}

	rawKey, prefix, verifier, err := s.generateKey()
	if err != nil {
		return nil, domain.NewInternal("could not generate API key")
	}
	key, err := s.repo.Create(ctx, repositories.APIKeyCreateInput{
		TenantID: tenantID, Name: name, Prefix: prefix, KeyVerifier: verifier,
		ProjectIDs: projectIDs, DefaultProjectID: req.DefaultProjectID, Scopes: scopes,
		ExpiresAt: req.ExpiresAt, CreatedBy: actor,
	})
	if err != nil {
		return nil, err
	}
	s.auditWrite(ctx, tenantID, actor, entities.AuditActionAPIKeyCreated, key, "created", nil)
	return &dtos.CreateAPIKeyResponse{Key: rawKey, APIKey: toAPIKeyDTO(*key)}, nil
}

// List returns non-secret metadata for all credentials owned by the tenant.
func (s *APIKeyService) List(ctx context.Context, tenantID string) ([]dtos.APIKeyDTO, error) {
	keys, err := s.repo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	result := make([]dtos.APIKeyDTO, 0, len(keys))
	for _, key := range keys {
		result = append(result, toAPIKeyDTO(key))
	}
	return result, nil
}

// Revoke disables a tenant-owned key for all subsequent MCP requests.
func (s *APIKeyService) Revoke(ctx context.Context, tenantID, actor, keyID string) (*dtos.APIKeyDTO, error) {
	key, err := s.repo.Revoke(ctx, tenantID, keyID)
	if err != nil {
		return nil, err
	}
	s.auditWrite(ctx, tenantID, actor, entities.AuditActionAPIKeyRevoked, key, "revoked", nil)
	result := toAPIKeyDTO(*key)
	return &result, nil
}

// Authenticate verifies raw key material and resolves a machine principal.
func (s *APIKeyService) Authenticate(ctx context.Context, rawKey string) (*entities.APIKeyPrincipal, error) {
	prefix, err := apiKeyPrefix(rawKey)
	if err != nil {
		return nil, domain.NewUnauthorized("invalid API key")
	}
	key, err := s.repo.FindByPrefix(ctx, prefix)
	if err != nil || !hmac.Equal(s.verifier(rawKey), key.KeyVerifier) {
		return nil, domain.NewUnauthorized("invalid API key")
	}
	if key.RevokedAt != nil {
		s.auditWrite(ctx, key.TenantID, key.Prefix, entities.AuditActionAPIKeyDenied, key, "revoked", nil)
		return nil, domain.NewUnauthorized("invalid API key")
	}
	if key.ExpiresAt != nil && !key.ExpiresAt.After(s.now()) {
		s.auditWrite(ctx, key.TenantID, key.Prefix, entities.AuditActionAPIKeyDenied, key, "expired", nil)
		return nil, domain.NewUnauthorized("invalid API key")
	}
	if err := s.repo.TouchLastUsed(ctx, key.ID); err != nil {
		return nil, err
	}
	s.auditWrite(ctx, key.TenantID, key.Prefix, entities.AuditActionAPIKeyUsed, key, "authenticated", nil)
	return &entities.APIKeyPrincipal{
		KeyID: key.ID, TenantID: key.TenantID, KeyPrefix: key.Prefix,
		ProjectIDs: append([]string(nil), key.ProjectIDs...), DefaultProjectID: key.DefaultProjectID,
		Scopes: append([]entities.MCPAPIScope(nil), key.Scopes...),
	}, nil
}

// RecordDeniedToolAccess records a safe authorization-denial audit entry.
func (s *APIKeyService) RecordDeniedToolAccess(ctx context.Context, principal entities.APIKeyPrincipal, reason string) {
	key := &entities.APIKey{ID: principal.KeyID, Prefix: principal.KeyPrefix}
	s.auditWrite(ctx, principal.TenantID, principal.KeyPrefix, entities.AuditActionAPIKeyDenied, key, reason, nil)
}

func (s *APIKeyService) generateKey() (rawKey, prefix string, verifier []byte, err error) {
	prefixBytes := make([]byte, 6)
	secretBytes := make([]byte, 32)
	if _, err = rand.Read(prefixBytes); err != nil {
		return "", "", nil, err
	}
	if _, err = rand.Read(secretBytes); err != nil {
		return "", "", nil, err
	}
	prefix = fmt.Sprintf("%s_%s", apiKeyPrefixMarker, hex.EncodeToString(prefixBytes))
	rawKey = prefix + "_" + base64.RawURLEncoding.EncodeToString(secretBytes)
	return rawKey, prefix, s.verifier(rawKey), nil
}

func (s *APIKeyService) verifier(rawKey string) []byte {
	mac := hmac.New(sha256.New, s.pepper)
	_, _ = mac.Write([]byte(rawKey))
	return mac.Sum(nil)
}

func apiKeyPrefix(rawKey string) (string, error) {
	parts := strings.SplitN(rawKey, "_", 3)
	if len(parts) != 3 || parts[0] != apiKeyPrefixMarker || parts[1] == "" || parts[2] == "" {
		return "", fmt.Errorf("malformed API key")
	}
	return parts[0] + "_" + parts[1], nil
}

func parseScopes(values []string) ([]entities.MCPAPIScope, error) {
	unique, err := uniqueNonEmpty(values)
	if err != nil {
		return nil, err
	}
	if len(unique) == 0 {
		return nil, domain.NewValidation("at least one scope is required")
	}
	scopes := make([]entities.MCPAPIScope, 0, len(unique))
	for _, value := range unique {
		scope := entities.MCPAPIScope(value)
		if _, ok := entities.ValidMCPAPIScopes[scope]; !ok {
			return nil, domain.NewValidation("unsupported API key scope")
		}
		scopes = append(scopes, scope)
	}
	return scopes, nil
}

func uniqueNonEmpty(values []string) ([]string, error) {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			return nil, domain.NewValidation("API key values cannot be empty")
		}
		if _, exists := seen[value]; exists {
			return nil, domain.NewValidation("API key values must be unique")
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result, nil
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func (s *APIKeyService) auditWrite(ctx context.Context, tenantID, actor string, action entities.AuditAction, key *entities.APIKey, status string, extra map[string]string) {
	if s.audit == nil {
		return
	}
	after := map[string]string{"prefix": key.Prefix, "status": status}
	for name, value := range extra {
		after[name] = value
	}
	s.audit.Write(ctx, tenantID, actor, action, "api_key", key.ID, nil, after, nil)
}

func toAPIKeyDTO(key entities.APIKey) dtos.APIKeyDTO {
	scopes := make([]string, 0, len(key.Scopes))
	for _, scope := range key.Scopes {
		scopes = append(scopes, string(scope))
	}
	return dtos.APIKeyDTO{
		ID: key.ID, TenantID: key.TenantID, Name: key.Name, Prefix: key.Prefix,
		ProjectIDs: append([]string(nil), key.ProjectIDs...), DefaultProjectID: key.DefaultProjectID,
		Scopes: scopes, ExpiresAt: key.ExpiresAt, RevokedAt: key.RevokedAt, LastUsedAt: key.LastUsedAt,
		CreatedBy: key.CreatedBy, CreatedAt: key.CreatedAt,
	}
}
