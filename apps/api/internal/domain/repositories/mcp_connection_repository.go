package repositories

import (
	"context"
	"time"

	"github.com/irosadie/open-state/api/internal/domain/entities"
)

// IMCPConnectionRepository persists project-owned MCP connection metadata.
// Every operation carries tenant and project scope so a connection from another
// project is indistinguishable from not found at the repository boundary.
type IMCPConnectionRepository interface {
	Create(ctx context.Context, input MCPConnectionCreateInput) (*entities.MCPConnection, error)
	FindByID(ctx context.Context, tenantID, projectID, id string) (*entities.MCPConnection, error)
	ListByProject(ctx context.Context, tenantID, projectID string) ([]entities.MCPConnection, error)
	Update(ctx context.Context, input MCPConnectionUpdateInput) (*entities.MCPConnection, error)
	Delete(ctx context.Context, tenantID, projectID, id string) error
	UpdateStatus(ctx context.Context, tenantID, projectID, id string, status entities.MCPConnectionStatus, actor string) (*entities.MCPConnection, error)
	RecordTest(ctx context.Context, tenantID, projectID, id string, status entities.MCPConnectionTestStatus, errorCode, actor string) (*entities.MCPConnection, error)
}

// IMCPConnectionSecurityRepository is an optional Phase 7 extension. Keeping
// it separate preserves compatibility with lightweight adapters used by older
// integrations while the PostgreSQL adapter implements the full lifecycle.
type IMCPConnectionSecurityRepository interface {
	IMCPConnectionRepository
	UpdateCredential(ctx context.Context, input MCPConnectionCredentialUpdateInput) (*entities.MCPConnection, error)
	UpdateOAuth(ctx context.Context, input MCPConnectionOAuthUpdateInput) (*entities.MCPConnection, error)
	DisconnectOAuth(ctx context.Context, tenantID, projectID, id, actor string) (*entities.MCPConnection, error)
	RecordHealth(ctx context.Context, input MCPConnectionHealthUpdateInput) (*entities.MCPConnection, error)
	ResetHealth(ctx context.Context, tenantID, projectID, id, actor string) (*entities.MCPConnection, error)
}

type MCPConnectionCredentialUpdateInput struct {
	TenantID            string
	ProjectID           string
	ID                  string
	CredentialReference *string
	CredentialStatus    entities.MCPConnectionCredentialStatus
	UpdatedBy           string
}

type MCPConnectionCreateInput struct {
	TenantID                   string
	ProjectID                  string
	Name                       string
	Alias                      string
	Transport                  entities.MCPConnectionTransport
	Endpoint                   *string
	StdioProfile               *string
	StdioArgs                  []string
	AuthType                   entities.MCPConnectionAuthType
	CredentialReference        *string
	CredentialStatus           entities.MCPConnectionCredentialStatus
	OAuthAuthorizationEndpoint *string
	OAuthTokenEndpoint         *string
	OAuthClientID              *string
	OAuthClientSecretReference *string
	OAuthScopes                []string
	OAuthRedirectURI           *string
	TimeoutMS                  int
	MaxConcurrency             int
	RateLimitPerSecond         float64
	RateLimitBurst             int
	RetryMax                   int
	CircuitFailureThreshold    int
	CircuitRecoverySeconds     int
	Status                     entities.MCPConnectionStatus
	CreatedBy                  string
}

type MCPConnectionUpdateInput struct {
	TenantID                   string
	ProjectID                  string
	ID                         string
	Name                       string
	Alias                      string
	Transport                  entities.MCPConnectionTransport
	Endpoint                   *string
	StdioProfile               *string
	StdioArgs                  []string
	AuthType                   entities.MCPConnectionAuthType
	CredentialReference        *string
	CredentialStatus           entities.MCPConnectionCredentialStatus
	OAuthAuthorizationEndpoint *string
	OAuthTokenEndpoint         *string
	OAuthClientID              *string
	OAuthClientSecretReference *string
	OAuthScopes                []string
	OAuthRedirectURI           *string
	TimeoutMS                  int
	MaxConcurrency             int
	RateLimitPerSecond         float64
	RateLimitBurst             int
	RetryMax                   int
	CircuitFailureThreshold    int
	CircuitRecoverySeconds     int
	UpdatedBy                  string
}

type MCPConnectionOAuthUpdateInput struct {
	TenantID              string
	ProjectID             string
	ID                    string
	CredentialReference   *string
	RefreshTokenReference *string
	CredentialStatus      entities.MCPConnectionCredentialStatus
	OAuthStatus           entities.MCPOAuthStatus
	OAuthExpiresAt        *time.Time
	UpdatedBy             string
}

type MCPConnectionHealthUpdateInput struct {
	TenantID            string
	ProjectID           string
	ID                  string
	HealthStatus        entities.MCPConnectionHealthStatus
	HealthReason        *string
	LastSuccessAt       *time.Time
	ConsecutiveFailures int
	CircuitOpenedAt     *time.Time
	Actor               string
}
