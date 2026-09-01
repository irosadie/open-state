package entities

import "time"

// MCPConnectionTransport identifies how OpenState reaches an external MCP server.
type MCPConnectionTransport string

const (
	MCPTransportStreamableHTTP MCPConnectionTransport = "streamable_http"
	MCPTransportSSE            MCPConnectionTransport = "sse"
	MCPTransportSTDIO          MCPConnectionTransport = "stdio"
)

// MCPConnectionAuthType identifies the authentication descriptor for a connection.
type MCPConnectionAuthType string

const (
	MCPAuthNone   MCPConnectionAuthType = "none"
	MCPAuthBearer MCPConnectionAuthType = "bearer"
	MCPAuthOAuth  MCPConnectionAuthType = "oauth"
)

type MCPConnectionStatus string

const (
	MCPConnectionEnabled  MCPConnectionStatus = "enabled"
	MCPConnectionDisabled MCPConnectionStatus = "disabled"
)

type MCPConnectionCredentialStatus string

const (
	MCPCredentialConfigured     MCPConnectionCredentialStatus = "configured"
	MCPCredentialMissing        MCPConnectionCredentialStatus = "missing"
	MCPCredentialActionRequired MCPConnectionCredentialStatus = "action_required"
)

type MCPConnectionTestStatus string

const (
	MCPTestNever    MCPConnectionTestStatus = "never"
	MCPTestReady    MCPConnectionTestStatus = "ready"
	MCPTestFailed   MCPConnectionTestStatus = "failed"
	MCPTestDisabled MCPConnectionTestStatus = "disabled"
)

// MCPConnection is a project-owned external MCP connection. Credential values
// are intentionally absent; only a protected reference is retained.
type MCPConnection struct {
	ID                  string
	TenantID            string
	ProjectID           string
	Name                string
	Alias               string
	Transport           MCPConnectionTransport
	Endpoint            *string
	StdioProfile        *string
	StdioArgs           []string
	AuthType            MCPConnectionAuthType
	CredentialReference *string
	CredentialStatus    MCPConnectionCredentialStatus
	Status              MCPConnectionStatus
	LastTestStatus      MCPConnectionTestStatus
	LastTestErrorCode   *string
	LastTestedAt        *time.Time
	CreatedBy           string
	UpdatedBy           string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}
