package entities

import (
	"encoding/json"
	"time"
)

// CapabilityEvidenceStatus is the outcome reported by the LLM host after it
// invokes a configured provider MCP tool.
type CapabilityEvidenceStatus string

const (
	CapabilityEvidenceSucceeded CapabilityEvidenceStatus = "SUCCEEDED"
	CapabilityEvidenceFailed    CapabilityEvidenceStatus = "FAILED"
)

// CapabilityExecutionEvidence is the explicit State MCP enforcement marker.
// It contains only normalized provider metadata and result data; credentials
// and connection headers never cross this boundary.
type CapabilityExecutionEvidence struct {
	ID                 string
	TenantID           string
	ProjectID          string
	WorkflowInstanceID string
	StateID            string
	CapabilityID       string
	CapabilityName     string
	ProviderServer     string
	ProviderTool       string
	CorrelationID      *string
	IdempotencyKey     string
	Status             CapabilityEvidenceStatus
	Result             json.RawMessage
	Error              json.RawMessage
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
