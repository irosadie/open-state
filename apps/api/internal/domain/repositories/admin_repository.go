package repositories

import (
	"context"

	"github.com/irosadie/open-state/api/internal/domain/entities"
)

// IAdminRepository is the persistence contract for current-tenant profile and
// membership administration. Every operation is scoped by tenantID.
type IAdminRepository interface {
	FindTenantByID(ctx context.Context, tenantID string) (*entities.Tenant, error)
	UpdateTenantProfile(ctx context.Context, tenantID, name, slug, description string) (*entities.Tenant, error)
	ListMemberships(ctx context.Context, tenantID string, search *string, offset, limit int) ([]entities.TenantMembership, error)
	CountMemberships(ctx context.Context, tenantID string, search *string) (int64, error)
	FindMembership(ctx context.Context, tenantID, userID string) (*entities.TenantMembership, error)
	CountOwners(ctx context.Context, tenantID string) (int64, error)
	AssignMembershipRole(ctx context.Context, tenantID, userID string, role entities.UserRole) (*entities.TenantMembership, error)
	RemoveMembership(ctx context.Context, tenantID, userID string) error

	// WithTx executes membership changes against one database transaction so the
	// last-Owner invariant is checked and written atomically.
	WithTx(ctx context.Context, fn func(IAdminRepository) error) error
}
