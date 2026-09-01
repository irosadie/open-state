package repositories

import (
	"context"

	"github.com/irosadie/open-state/api/internal/domain/entities"
)

// ICapabilityRepository defines the persistence contract for the Capability
// Registry, scoped capability bindings, and policies. It is tenant-scoped:
// every method takes an explicit tenantID (PRD §4, §96) so cross-tenant access
// is impossible at the data-access layer. It operates on domain entities
// (DB-agnostic, ADR-001) and returns NOT_FOUND/CONFLICT DomainErrors.
type ICapabilityRepository interface {
	// Create persists a new capability in the registry (PRD §59).
	Create(ctx context.Context, tenantID, name string, description *string, providerType entities.ProviderType, providerID, providerTool *string, inputSchema, outputSchema []byte, version int, credentialReference *string) (*entities.Capability, error)
	// FindByID returns a capability by id within a tenant.
	FindByID(ctx context.Context, tenantID, id string) (*entities.Capability, error)
	// FindByName returns a capability by logical name within a tenant.
	FindByName(ctx context.Context, tenantID, name string) (*entities.Capability, error)
	// ListByTenant returns all capabilities for a tenant.
	ListByTenant(ctx context.Context, tenantID string) ([]entities.Capability, error)
	// ListByTenantFiltered returns capabilities for a tenant, optionally
	// filtered by provider type and status (empty values = no filter).
	ListByTenantFiltered(ctx context.Context, tenantID string, providerType entities.ProviderType, capStatus entities.CapabilityStatus) ([]entities.Capability, error)
	// Update updates mutable fields of a capability within a tenant.
	Update(ctx context.Context, tenantID, id string, description *string, providerType entities.ProviderType, providerID, providerTool *string, inputSchema, outputSchema []byte, status entities.CapabilityStatus, version int, credentialReference *string) (*entities.Capability, error)
	// UpdateStatus updates a capability's status within a tenant.
	UpdateStatus(ctx context.Context, tenantID, id string, status entities.CapabilityStatus) (*entities.Capability, error)
	// Disable marks a capability DISABLED within a tenant.
	Disable(ctx context.Context, tenantID, id string) (*entities.Capability, error)

	// Bind scopes a capability's availability to a tenant/workflow/state level
	// with most-restrictive-wins resolution (PRD §60).
	Bind(ctx context.Context, tenantID, capabilityID string, scopeType entities.BindingScopeType, scopeID string, permission entities.BindingPermission) (*entities.CapabilityBinding, error)
	// ListBindingsByCapability returns all bindings for a capability (resolution inputs).
	ListBindingsByCapability(ctx context.Context, tenantID, capabilityID string) ([]entities.CapabilityBinding, error)
	// ListBindingsByScope returns all capability bindings at a given scope.
	ListBindingsByScope(ctx context.Context, tenantID string, scopeType entities.BindingScopeType, scopeID string) ([]entities.CapabilityBinding, error)
	// Unbind deletes a capability binding within a tenant.
	Unbind(ctx context.Context, tenantID, bindingID string) error

	// UpsertPolicy inserts or replaces a policy for a scope (PRD §3.13, §12).
	UpsertPolicy(ctx context.Context, tenantID string, scopeType entities.PolicyScopeType, scopeID, policyType string, content []byte) (*entities.Policy, error)
	// FindPolicyByType returns a policy by its type within a scope.
	FindPolicyByType(ctx context.Context, tenantID string, scopeType entities.PolicyScopeType, scopeID, policyType string) (*entities.Policy, error)
	// ListPoliciesByScope returns all policies for a scope.
	ListPoliciesByScope(ctx context.Context, tenantID string, scopeType entities.PolicyScopeType, scopeID string) ([]entities.Policy, error)
}
