package dtos

import "time"

// CreateAPIKeyRequest describes a tenant-scoped State MCP machine credential.
type CreateAPIKeyRequest struct {
	Name             string     `json:"name"`
	ProjectIDs       []string   `json:"projectIds"`
	DefaultProjectID *string    `json:"defaultProjectId"`
	Scopes           []string   `json:"scopes"`
	ExpiresAt        *time.Time `json:"expiresAt"`
}

// APIKeyDTO is safe to return after creation and from list endpoints. It never
// includes key verifier material or the raw API key secret.
type APIKeyDTO struct {
	ID               string     `json:"id"`
	TenantID         string     `json:"tenantId"`
	Name             string     `json:"name"`
	Prefix           string     `json:"prefix"`
	ProjectIDs       []string   `json:"projectIds"`
	DefaultProjectID *string    `json:"defaultProjectId"`
	Scopes           []string   `json:"scopes"`
	ExpiresAt        *time.Time `json:"expiresAt"`
	RevokedAt        *time.Time `json:"revokedAt"`
	LastUsedAt       *time.Time `json:"lastUsedAt"`
	CreatedBy        string     `json:"createdBy"`
	CreatedAt        time.Time  `json:"createdAt"`
}

// CreateAPIKeyResponse is the only response type that carries a raw API key.
// Clients must store Key immediately because it is not retrievable afterward.
type CreateAPIKeyResponse struct {
	Key    string    `json:"key"`
	APIKey APIKeyDTO `json:"apiKey"`
}
