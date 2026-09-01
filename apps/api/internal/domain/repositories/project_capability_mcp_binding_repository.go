package repositories

import (
	"context"

	"github.com/irosadie/open-state/api/internal/domain/entities"
)

// IProjectCapabilityMCPBindingRepository persists explicit project-scoped
// mappings from logical MCP capabilities to discovered provider tools.
type IProjectCapabilityMCPBindingRepository interface {
	ListEligibleToolOptions(ctx context.Context, tenantID, projectID string) ([]entities.ProjectMCPToolOption, error)
	ListByProject(ctx context.Context, tenantID, projectID string) ([]entities.ProjectCapabilityMCPBinding, error)
	FindByCapability(ctx context.Context, tenantID, projectID, capabilityID string) (*entities.ProjectCapabilityMCPBinding, error)
	Upsert(ctx context.Context, input ProjectCapabilityMCPBindingUpsertInput) error
	Delete(ctx context.Context, tenantID, projectID, capabilityID string) error
}

type ProjectCapabilityMCPBindingUpsertInput struct {
	TenantID        string
	ProjectID       string
	CapabilityID    string
	ConnectionID    string
	ToolID          string
	ToolFingerprint string
}
