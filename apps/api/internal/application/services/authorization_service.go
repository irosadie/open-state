package services

import (
	"context"

	"github.com/irosadie/open-state/api/internal/domain/entities"
	domainsvc "github.com/irosadie/open-state/api/internal/domain/services"
	domain "github.com/irosadie/open-state/go-shared/domain"
)

// AuthorizationService resolves a user's tenant role and checks permissions
// against the domain role-permission matrix (PRD 80, 81). It is the application
// seam used by the HTTP RequirePermission middleware and the capability
// invocation chain. Default-deny: an absent role assignment yields an empty
// permission set.
type AuthorizationService struct {
	roles domainRoleRepo
}

// Permission is the domain permission type re-exported for the application layer.
type Permission = domainsvc.Permission

// domainRoleRepo is the minimal role-assignment persistence contract consumed by
// the authorization service. It is satisfied by repositories.IRoleAssignmentRepository.
type domainRoleRepo interface {
	FindRoleByUserAndTenant(ctx context.Context, userID, tenantID string) (entities.UserRole, error)
}

// NewAuthorizationService builds an AuthorizationService over the role-assignment
// repository.
func NewAuthorizationService(roles domainRoleRepo) *AuthorizationService {
	return &AuthorizationService{roles: roles}
}

// RoleForTenant returns the user's effective role for a tenant. An absent role
// assignment resolves to the least-privilege VIEWER (default deny of elevated
// permissions).
func (s *AuthorizationService) RoleForTenant(ctx context.Context, userID, tenantID string) (entities.UserRole, error) {
	role, err := s.roles.FindRoleByUserAndTenant(ctx, userID, tenantID)
	if err != nil {
		return "", err
	}
	if role == "" {
		return entities.UserRoleViewer, nil
	}
	return role, nil
}

// PermissionsForTenant returns the user's granted permissions for a tenant,
// derived from their effective role and the domain matrix.
func (s *AuthorizationService) PermissionsForTenant(ctx context.Context, userID, tenantID string) ([]Permission, error) {
	role, err := s.RoleForTenant(ctx, userID, tenantID)
	if err != nil {
		return nil, err
	}
	return domainsvc.PermissionsForRole(role), nil
}

// Authorize reports whether the user holds the required permission for the
// tenant. If the user is not assigned a role, authorization is denied.
func (s *AuthorizationService) Authorize(ctx context.Context, userID, tenantID string, required Permission) (bool, error) {
	role, err := s.RoleForTenant(ctx, userID, tenantID)
	if err != nil {
		return false, err
	}
	return domainsvc.HasPermission(role, required), nil
}

// Require returns nil if the user is authorized, or a FORBIDDEN domain error.
func (s *AuthorizationService) Require(ctx context.Context, userID, tenantID string, required Permission) error {
	ok, err := s.Authorize(ctx, userID, tenantID, required)
	if err != nil {
		return err
	}
	if !ok {
		return domain.NewForbidden("permission denied")
	}
	return nil
}
