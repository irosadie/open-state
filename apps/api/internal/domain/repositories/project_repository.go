package repositories

import (
	"context"

	"github.com/irosadie/open-state/api/internal/domain/entities"
)

// IProjectRepository defines the persistence contract for projects (business
// areas owned by a tenant, PRD §3.1.1). Tenant-scoped: every method takes an
// explicit tenantID (PRD §4, §96). Operates on domain entities (ADR-001).
type IProjectRepository interface {
	// Create persists a new project within a tenant.
	Create(ctx context.Context, tenantID, name, slug string, status entities.ProjectStatus) (*entities.Project, error)
	// FindByID returns a project by id within a tenant.
	FindByID(ctx context.Context, tenantID, id string) (*entities.Project, error)
	// FindBySlug returns a project by slug within a tenant.
	FindBySlug(ctx context.Context, tenantID, slug string) (*entities.Project, error)
	// ListByTenant returns all projects for a tenant.
	ListByTenant(ctx context.Context, tenantID string) ([]entities.Project, error)
}
