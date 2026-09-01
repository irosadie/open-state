package repositories

import (
	"context"
	"encoding/json"

	"github.com/irosadie/open-state/api/internal/domain/entities"
)

// IMCPToolCatalogRepository stores the current project-scoped tool snapshot
// while retaining discovery-run outcomes for safe operational visibility.
type IMCPToolCatalogRepository interface {
	Get(ctx context.Context, tenantID, projectID, connectionID string) (*entities.MCPToolCatalog, error)
	Reconcile(ctx context.Context, input MCPToolCatalogReconcileInput) (*entities.MCPDiscoveryRun, error)
	RecordFailure(ctx context.Context, input MCPToolCatalogFailureInput) (*entities.MCPDiscoveryRun, error)
	SetEnabled(ctx context.Context, tenantID, projectID, connectionID, toolName string, enabled bool) (*entities.MCPDiscoveredTool, error)
}

type MCPDiscoveredToolInput struct {
	Name        string
	Title       string
	Description string
	InputSchema json.RawMessage
	Annotations json.RawMessage
	Fingerprint string
}

type MCPToolCatalogReconcileInput struct {
	TenantID           string
	ProjectID          string
	ConnectionID       string
	Actor              string
	CatalogFingerprint string
	Tools              []MCPDiscoveredToolInput
}

type MCPToolCatalogFailureInput struct {
	TenantID     string
	ProjectID    string
	ConnectionID string
	Actor        string
	ErrorCode    string
}
