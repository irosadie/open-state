package repositories

import (
	"context"

	"github.com/irosadie/open-state/api/internal/domain/entities"
)

// IIntentRepository defines persistence for canonical, tenant/project-scoped
// intent mappings. Routable reads only return mappings to published workflows.
type IIntentRepository interface {
	// ListRoutable returns intents for a project whose workflow is published.
	ListRoutable(ctx context.Context, tenantID, projectID string) ([]entities.Intent, error)
	// FindRoutable returns a canonical intent key if it maps to a published workflow.
	FindRoutable(ctx context.Context, tenantID, projectID, key string) (*entities.Intent, error)
	// Upsert creates or updates a seed/provisioned intent mapping.
	Upsert(ctx context.Context, tenantID, projectID, workflowID, key, name, description string, examples []string) (*entities.Intent, error)
}
