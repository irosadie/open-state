package repositories

import (
	"context"

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

type MCPConnectionCreateInput struct {
	TenantID            string
	ProjectID           string
	Name                string
	Alias               string
	Transport           entities.MCPConnectionTransport
	Endpoint            *string
	StdioProfile        *string
	StdioArgs           []string
	AuthType            entities.MCPConnectionAuthType
	CredentialReference *string
	CredentialStatus    entities.MCPConnectionCredentialStatus
	Status              entities.MCPConnectionStatus
	CreatedBy           string
}

type MCPConnectionUpdateInput struct {
	TenantID            string
	ProjectID           string
	ID                  string
	Name                string
	Alias               string
	Transport           entities.MCPConnectionTransport
	Endpoint            *string
	StdioProfile        *string
	StdioArgs           []string
	AuthType            entities.MCPConnectionAuthType
	CredentialReference *string
	CredentialStatus    entities.MCPConnectionCredentialStatus
	UpdatedBy           string
}
