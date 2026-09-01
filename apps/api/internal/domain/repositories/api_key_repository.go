package repositories

import (
	"context"
	"time"

	"github.com/irosadie/open-state/api/internal/domain/entities"
)

// APIKeyCreateInput contains the persisted, non-secret properties of a new
// machine credential. KeyVerifier is a one-way keyed verifier, never raw key
// material.
type APIKeyCreateInput struct {
	TenantID         string
	Name             string
	Prefix           string
	KeyVerifier      []byte
	ProjectIDs       []string
	DefaultProjectID *string
	Scopes           []entities.MCPAPIScope
	ExpiresAt        *time.Time
	CreatedBy        string
}

// IAPIKeyRepository owns persistence for State MCP machine credentials.
type IAPIKeyRepository interface {
	Create(ctx context.Context, input APIKeyCreateInput) (*entities.APIKey, error)
	FindByPrefix(ctx context.Context, prefix string) (*entities.APIKey, error)
	ListByTenant(ctx context.Context, tenantID string) ([]entities.APIKey, error)
	Revoke(ctx context.Context, tenantID, keyID string) (*entities.APIKey, error)
	TouchLastUsed(ctx context.Context, keyID string) error
}
