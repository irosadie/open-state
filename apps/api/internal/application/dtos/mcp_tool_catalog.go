package dtos

import "encoding/json"

type MCPDiscoveryRunDTO struct {
	ID                 string  `json:"id"`
	TenantID           string  `json:"tenantId"`
	ProjectID          string  `json:"projectId"`
	ConnectionID       string  `json:"connectionId"`
	Status             string  `json:"status"`
	ToolCount          int     `json:"toolCount"`
	CatalogFingerprint *string `json:"catalogFingerprint"`
	ErrorCode          *string `json:"errorCode"`
	StartedAt          string  `json:"startedAt"`
	CompletedAt        string  `json:"completedAt"`
	CreatedBy          string  `json:"createdBy"`
}

// MCPDiscoveredToolDTO contains sanitized provider metadata only. It never
// includes the external connection endpoint, headers, or credential reference.
type MCPDiscoveredToolDTO struct {
	ID             string          `json:"id"`
	TenantID       string          `json:"tenantId"`
	ProjectID      string          `json:"projectId"`
	ConnectionID   string          `json:"connectionId"`
	Name           string          `json:"name"`
	Title          *string         `json:"title"`
	Description    string          `json:"description"`
	InputSchema    json.RawMessage `json:"inputSchema"`
	Annotations    json.RawMessage `json:"annotations"`
	Fingerprint    string          `json:"fingerprint"`
	Enabled        bool            `json:"enabled"`
	Availability   string          `json:"availability"`
	DriftStatus    string          `json:"driftStatus"`
	FirstSeenAt    string          `json:"firstSeenAt"`
	LastSeenAt     string          `json:"lastSeenAt"`
	RemovedAt      *string         `json:"removedAt"`
	DiscoveryRunID *string         `json:"discoveryRunId"`
	CreatedAt      string          `json:"createdAt"`
	UpdatedAt      string          `json:"updatedAt"`
}

type MCPToolCatalogDTO struct {
	ConnectionID      string                 `json:"connectionId"`
	Tools             []MCPDiscoveredToolDTO `json:"tools"`
	LatestRun         *MCPDiscoveryRunDTO    `json:"latestRun"`
	LastSuccessfulRun *MCPDiscoveryRunDTO    `json:"lastSuccessfulRun"`
}

type SetMCPToolEnabledRequest struct {
	Enabled *bool `json:"enabled"`
}
