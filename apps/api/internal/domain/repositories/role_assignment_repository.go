package repositories

import (
	"context"

	"github.com/irosadie/open-state/api/internal/domain/entities"
)

// IRoleAssignmentRepository defines the persistence contract for tenant-scoped
// RBAC role assignments (PRD 80, 81). Every method takes an explicit tenantID
// (PRD 4, 96) so cross-tenant access is impossible at the data-access layer.
type IRoleAssignmentRepository interface {
	// Assign creates or updates the role assignment for the (user_id, tenant_id)
	// pair (upsert). Returns the resulting assignment.
	Assign(ctx context.Context, userID, tenantID string, role entities.UserRole) (*entities.RoleAssignment, error)
	// FindRoleByUserAndTenant returns the user's role for a tenant.
	FindRoleByUserAndTenant(ctx context.Context, userID, tenantID string) (entities.UserRole, error)
	// ListByTenant returns all role assignments in a tenant.
	ListByTenant(ctx context.Context, tenantID string) ([]entities.RoleAssignment, error)
	// Remove deletes a user's role assignment for a tenant.
	Remove(ctx context.Context, userID, tenantID string) error
}
