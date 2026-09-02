package services

import "context"

// SecretStore is the application boundary for protected credentials. Concrete
// deployments can bind it to an environment adapter, Vault, KMS, or another
// secret manager. Callers only persist opaque references and lifecycle status.
type SecretStore interface {
	Put(ctx context.Context, tenantID, projectID, kind, value string) (string, error)
	Resolve(ctx context.Context, tenantID, projectID, reference string) (string, error)
	Rotate(ctx context.Context, tenantID, projectID, reference, kind, value string) (string, error)
	Revoke(ctx context.Context, tenantID, projectID, reference string) error
	Status(ctx context.Context, tenantID, projectID, reference string) (SecretStatus, error)
}

type SecretStatus string

const (
	SecretStatusConfigured SecretStatus = "configured"
	SecretStatusMissing    SecretStatus = "missing"
	SecretStatusRevoked    SecretStatus = "revoked"
)

// OAuthAccessTokenProvider supplies a short-lived access token to the MCP
// transport. It may refresh the token server-side and must never return token
// material to an HTTP or MCP client.
type OAuthAccessTokenProvider interface {
	AccessToken(ctx context.Context, connectionID, tenantID, projectID string) (string, error)
}
