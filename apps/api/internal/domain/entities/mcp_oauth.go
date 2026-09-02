package entities

import "time"

type MCPAuthorizationTransactionStatus string

const (
	MCPAuthorizationPending  MCPAuthorizationTransactionStatus = "pending"
	MCPAuthorizationConsumed MCPAuthorizationTransactionStatus = "consumed"
	MCPAuthorizationExpired  MCPAuthorizationTransactionStatus = "expired"
)

// MCPAuthorizationTransaction stores only a state hash and a secret-store
// reference to the PKCE verifier. Raw OAuth state, verifier, code, and tokens
// are never persisted.
type MCPAuthorizationTransaction struct {
	ID                string
	TenantID          string
	ProjectID         string
	ConnectionID      string
	StateHash         []byte
	VerifierReference string
	RedirectURI       string
	ExpiresAt         time.Time
	Status            MCPAuthorizationTransactionStatus
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
