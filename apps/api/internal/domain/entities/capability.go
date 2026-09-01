package entities

import (
	"database/sql"
	"encoding/json"
	"time"
)

// ProviderType identifies the runtime provider that fulfills a capability
// (PRD §59). FUTURE is reserved for providers not yet integrated.
type ProviderType string

const (
	ProviderTypeMCP      ProviderType = "MCP"
	ProviderTypeInternal ProviderType = "INTERNAL"
	ProviderTypeHTTP     ProviderType = "HTTP"
	ProviderTypeFuture   ProviderType = "FUTURE"
)

// CapabilityStatus is the lifecycle of a capability in the registry.
type CapabilityStatus string

const (
	CapabilityActive   CapabilityStatus = "ACTIVE"
	CapabilityInactive CapabilityStatus = "INACTIVE"
	CapabilityDisabled CapabilityStatus = "DISABLED"
)

// Capability is a logical operation in the Capability Registry (PRD §3.11, §59),
// referenced by states and resolved to a provider at runtime. Secrets are never
// stored here — only a credential_reference (PRD §61).
type Capability struct {
	ID                  string
	TenantID            string
	Name                string // logical capability, e.g. payment.create
	Description         sql.NullString
	ProviderType        ProviderType
	ProviderID          sql.NullString  // stable provider MCP server alias
	ProviderTool        sql.NullString  // concrete provider MCP tool name
	InputSchema         json.RawMessage // JSON schema of required inputs
	OutputSchema        json.RawMessage // JSON schema of produced outputs
	Status              CapabilityStatus
	Version             int
	CredentialReference sql.NullString // PRD §61, never secrets
	CreatedAt           time.Time
	UpdatedAt           time.Time
}
