package entities

import (
	"encoding/json"
	"time"
)

// MCPDiscoveryRunStatus describes the outcome of one explicit tools/list refresh.
type MCPDiscoveryRunStatus string

const (
	MCPDiscoverySucceeded MCPDiscoveryRunStatus = "succeeded"
	MCPDiscoveryFailed    MCPDiscoveryRunStatus = "failed"
)

// MCPDiscoveredToolAvailability indicates whether the provider returned a tool
// in the latest successful catalog.
type MCPDiscoveredToolAvailability string

const (
	MCPToolPresent MCPDiscoveredToolAvailability = "present"
	MCPToolRemoved MCPDiscoveredToolAvailability = "removed"
)

// MCPDiscoveredToolDriftStatus records how the tool changed in the latest
// successful refresh. It is informational and never rewrites authored bindings.
type MCPDiscoveredToolDriftStatus string

const (
	MCPToolNew          MCPDiscoveredToolDriftStatus = "new"
	MCPToolUnchanged    MCPDiscoveredToolDriftStatus = "unchanged"
	MCPToolChanged      MCPDiscoveredToolDriftStatus = "changed"
	MCPToolDriftRemoved MCPDiscoveredToolDriftStatus = "removed"
)

type MCPDiscoveryRun struct {
	ID                 string
	TenantID           string
	ProjectID          string
	ConnectionID       string
	Status             MCPDiscoveryRunStatus
	ToolCount          int
	CatalogFingerprint *string
	ErrorCode          *string
	StartedAt          time.Time
	CompletedAt        time.Time
	CreatedBy          string
}

// MCPDiscoveredTool is sanitized provider metadata. It deliberately contains
// no endpoint, headers, credentials, or executable information.
type MCPDiscoveredTool struct {
	ID             string
	TenantID       string
	ProjectID      string
	ConnectionID   string
	Name           string
	Title          *string
	Description    string
	InputSchema    json.RawMessage
	Annotations    json.RawMessage
	Fingerprint    string
	Enabled        bool
	Availability   MCPDiscoveredToolAvailability
	DriftStatus    MCPDiscoveredToolDriftStatus
	FirstSeenAt    time.Time
	LastSeenAt     time.Time
	RemovedAt      *time.Time
	DiscoveryRunID *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type MCPToolCatalog struct {
	ConnectionID      string
	Tools             []MCPDiscoveredTool
	LatestRun         *MCPDiscoveryRun
	LastSuccessfulRun *MCPDiscoveryRun
}
