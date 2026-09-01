package entities

import (
	"encoding/json"
	"time"
)

// ProjectCapabilityMCPBindingHealth describes whether the explicit project
// binding can currently be used by State MCP.
type ProjectCapabilityMCPBindingHealth string

const (
	ProjectCapabilityMCPBindingActive             ProjectCapabilityMCPBindingHealth = "ACTIVE"
	ProjectCapabilityMCPBindingMissingMapping     ProjectCapabilityMCPBindingHealth = "MISSING_MAPPING"
	ProjectCapabilityMCPBindingConnectionDisabled ProjectCapabilityMCPBindingHealth = "CONNECTION_DISABLED"
	ProjectCapabilityMCPBindingToolDisabled       ProjectCapabilityMCPBindingHealth = "TOOL_DISABLED"
	ProjectCapabilityMCPBindingToolRemoved        ProjectCapabilityMCPBindingHealth = "TOOL_REMOVED"
	ProjectCapabilityMCPBindingStale              ProjectCapabilityMCPBindingHealth = "STALE"
)

// ProjectMCPToolOption is the safe catalog projection used by authoring. It
// contains no endpoint, headers, credential reference, or secret material.
type ProjectMCPToolOption struct {
	ConnectionID     string
	ConnectionName   string
	ConnectionAlias  string
	ConnectionStatus MCPConnectionStatus
	ToolID           string
	ToolName         string
	ToolTitle        *string
	ToolDescription  string
	InputSchema      json.RawMessage
	ToolFingerprint  string
}

// ProjectCapabilityMCPBinding is the logical capability to provider-tool
// mapping for one project, including safe health metadata for the builder and
// State MCP projection.
type ProjectCapabilityMCPBinding struct {
	ID                     string
	TenantID               string
	ProjectID              string
	CapabilityID           string
	CapabilityName         string
	CapabilityDescription  *string
	ConnectionID           string
	ConnectionName         string
	ConnectionAlias        string
	ConnectionStatus       MCPConnectionStatus
	ToolID                 string
	ToolName               string
	ToolTitle              *string
	ToolDescription        string
	BoundToolFingerprint   string
	CurrentToolFingerprint string
	ToolEnabled            bool
	ToolAvailability       MCPDiscoveredToolAvailability
	ToolDriftStatus        MCPDiscoveredToolDriftStatus
	Health                 ProjectCapabilityMCPBindingHealth
	HealthReason           string
	CreatedAt              time.Time
	UpdatedAt              time.Time
}
