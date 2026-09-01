package dtos

import "encoding/json"

// MCPToolOptionDTO is the safe provider catalog projection used by State
// Builder. Endpoint and credential fields are intentionally absent.
type MCPToolOptionDTO struct {
	ConnectionID     string          `json:"connectionId"`
	ConnectionName   string          `json:"connectionName"`
	ConnectionAlias  string          `json:"connectionAlias"`
	ConnectionStatus string          `json:"connectionStatus"`
	ToolID           string          `json:"toolId"`
	ToolName         string          `json:"toolName"`
	ToolTitle        *string         `json:"toolTitle"`
	ToolDescription  string          `json:"toolDescription"`
	InputSchema      json.RawMessage `json:"inputSchema"`
	ToolFingerprint  string          `json:"toolFingerprint"`
}

type MCPToolOptionListDTO struct {
	Data []MCPToolOptionDTO `json:"data"`
}

type ProjectCapabilityMCPBindingDTO struct {
	ID                     string  `json:"id"`
	TenantID               string  `json:"tenantId"`
	ProjectID              string  `json:"projectId"`
	CapabilityID           string  `json:"capabilityId"`
	CapabilityName         string  `json:"capabilityName"`
	CapabilityDescription  *string `json:"capabilityDescription"`
	ConnectionID           string  `json:"connectionId,omitempty"`
	ConnectionName         string  `json:"connectionName,omitempty"`
	ConnectionAlias        string  `json:"connectionAlias,omitempty"`
	ConnectionStatus       string  `json:"connectionStatus,omitempty"`
	ToolID                 string  `json:"toolId,omitempty"`
	ToolName               string  `json:"toolName,omitempty"`
	ToolTitle              *string `json:"toolTitle,omitempty"`
	ToolDescription        string  `json:"toolDescription,omitempty"`
	BoundToolFingerprint   string  `json:"boundToolFingerprint,omitempty"`
	CurrentToolFingerprint string  `json:"currentToolFingerprint,omitempty"`
	ToolEnabled            *bool   `json:"toolEnabled,omitempty"`
	ToolAvailability       string  `json:"toolAvailability,omitempty"`
	ToolDriftStatus        string  `json:"toolDriftStatus,omitempty"`
	Health                 string  `json:"health"`
	HealthReason           string  `json:"healthReason"`
	CreatedAt              string  `json:"createdAt,omitempty"`
	UpdatedAt              string  `json:"updatedAt,omitempty"`
}

type ProjectCapabilityMCPBindingListDTO struct {
	Data []ProjectCapabilityMCPBindingDTO `json:"data"`
}

// UpsertProjectCapabilityMCPBindingRequest accepts only catalog IDs. Raw
// endpoint URLs and arbitrary tool names are deliberately not accepted.
type UpsertProjectCapabilityMCPBindingRequest struct {
	ConnectionID string `json:"connectionId"`
	ToolID       string `json:"toolId"`
}
