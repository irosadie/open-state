package repositories

import (
	"context"

	"github.com/irosadie/open-state/api/internal/domain/entities"
)

// CapabilityEvidenceInput is the normalized, secret-safe report accepted by
// State MCP after a direct provider MCP invocation.
type CapabilityEvidenceInput struct {
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
	Status             entities.CapabilityEvidenceStatus
	Result             []byte
	Error              []byte
}

// ICapabilityEvidenceRepository stores the explicit evidence used by the
// state transition gate. All reads and writes are tenant/project scoped.
type ICapabilityEvidenceRepository interface {
	Upsert(ctx context.Context, input CapabilityEvidenceInput) (*entities.CapabilityExecutionEvidence, error)
	FindByIdempotency(ctx context.Context, tenantID, projectID, instanceID, stateID, capabilityID, idempotencyKey string) (*entities.CapabilityExecutionEvidence, error)
	ListByState(ctx context.Context, tenantID, projectID, instanceID, stateID string) ([]entities.CapabilityExecutionEvidence, error)
	ListByInstanceState(ctx context.Context, tenantID, instanceID, stateID string) ([]entities.CapabilityExecutionEvidence, error)
}
