package dtos

// Capability request/response DTOs for the admin API (PRD §59-64).
// Secrets are never exposed — only the credential_reference string (PRD §61, §91).

// CreateCapabilityRequest is the payload to register a capability.
type CreateCapabilityRequest struct {
	Name                string `json:"name"`
	Description         string `json:"description"`
	ProviderType        string `json:"providerType"`
	ProviderID          string `json:"providerId"`
	InputSchema         any    `json:"inputSchema"`
	OutputSchema        any    `json:"outputSchema"`
	Version             int    `json:"version"`
	CredentialReference string `json:"credentialReference"`
}

// UpdateCapabilityRequest is the payload to update a capability's mutable fields.
type UpdateCapabilityRequest struct {
	Description         string `json:"description"`
	ProviderType        string `json:"providerType"`
	ProviderID          string `json:"providerId"`
	InputSchema         any    `json:"inputSchema"`
	OutputSchema        any    `json:"outputSchema"`
	Status              string `json:"status"`
	Version             int    `json:"version"`
	CredentialReference string `json:"credentialReference"`
}

// CapabilityDTO is the serializable capability returned to callers. It exposes
// only credential_reference, never the resolved secret.
type CapabilityDTO struct {
	ID                  string  `json:"id"`
	TenantID            string  `json:"tenantId"`
	Name                string  `json:"name"`
	Description         *string `json:"description"`
	ProviderType        string  `json:"providerType"`
	ProviderID          *string `json:"providerId"`
	InputSchema         any     `json:"inputSchema"`
	OutputSchema        any     `json:"outputSchema"`
	Status              string  `json:"status"`
	Version             int     `json:"version"`
	CredentialReference *string `json:"credentialReference"`
	CreatedAt           string  `json:"createdAt"`
	UpdatedAt           string  `json:"updatedAt"`
}

// CapabilityListDTO wraps a tenant-scoped capability list.
type CapabilityListDTO struct {
	Data []CapabilityDTO `json:"data"`
}

// CreateBindingRequest binds a capability to a tenant/workflow/state scope.
type CreateBindingRequest struct {
	ScopeType  string `json:"scopeType"`
	ScopeID    string `json:"scopeId"`
	Permission string `json:"permission"`
}

// CapabilityBindingDTO is the serializable capability binding.
type CapabilityBindingDTO struct {
	ID           string `json:"id"`
	TenantID     string `json:"tenantId"`
	CapabilityID string `json:"capabilityId"`
	ScopeType    string `json:"scopeType"`
	ScopeID      string `json:"scopeId"`
	Permission   string `json:"permission"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

// TestInvocationRequest is the payload to test-invoke a capability in
// sandbox/mock mode (PRD §2064).
type TestInvocationRequest struct {
	Payload map[string]any `json:"payload"`
	ScopeID string         `json:"scopeId"` // optional workflow/state scope for binding resolution
}

// TestInvocationResultDTO is the normalized result flagged fromMock.
type TestInvocationResultDTO struct {
	Data       map[string]any `json:"data"`
	FromMock   bool           `json:"fromMock"`
	DurationMS int64          `json:"durationMs"`
	Event      *string        `json:"event"`
}

// CapabilityErrorDTO is a classified capability failure (PRD §87).
type CapabilityErrorDTO struct {
	Kind    string `json:"kind"`
	Code    string `json:"code"`
	Message string `json:"message"`
}
