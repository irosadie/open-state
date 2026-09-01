package entities

import "time"

// MCPAPIScope grants a machine principal access to a State MCP tool category.
type MCPAPIScope string

const (
	MCPAPIScopeStateRead        MCPAPIScope = "state:read"
	MCPAPIScopeStateWrite       MCPAPIScope = "state:write"
	MCPAPIScopeCapabilityInvoke MCPAPIScope = "capability:invoke"
)

// ValidMCPAPIScopes is the complete allowlist for machine API key scopes.
var ValidMCPAPIScopes = map[MCPAPIScope]struct{}{
	MCPAPIScopeStateRead:        {},
	MCPAPIScopeStateWrite:       {},
	MCPAPIScopeCapabilityInvoke: {},
}

// APIKey is the non-secret, persisted representation of a machine credential.
// KeyVerifier is retained for authentication but is never exposed through DTOs.
type APIKey struct {
	ID               string
	TenantID         string
	Name             string
	Prefix           string
	KeyVerifier      []byte
	ProjectIDs       []string
	DefaultProjectID *string
	Scopes           []MCPAPIScope
	ExpiresAt        *time.Time
	RevokedAt        *time.Time
	LastUsedAt       *time.Time
	CreatedBy        string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// APIKeyPrincipal is the immutable authenticated identity available to State
// MCP handlers. It deliberately excludes the verifier and raw key secret.
type APIKeyPrincipal struct {
	KeyID            string
	TenantID         string
	KeyPrefix        string
	ProjectIDs       []string
	DefaultProjectID *string
	Scopes           []MCPAPIScope
}

// HasScope reports whether the principal is allowed to use a scope.
func (p APIKeyPrincipal) HasScope(scope MCPAPIScope) bool {
	for _, granted := range p.Scopes {
		if granted == scope {
			return true
		}
	}
	return false
}

// AllowsProject reports whether a project belongs to this machine principal's
// explicit allowlist.
func (p APIKeyPrincipal) AllowsProject(projectID string) bool {
	for _, allowed := range p.ProjectIDs {
		if allowed == projectID {
			return true
		}
	}
	return false
}
