package capability

import (
	"context"
	"errors"

	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	"github.com/irosadie/open-state/go-shared/domain"
)

// ResolvedCapability is the output of resolving a logical capability to a
// provider + schema + bindings (PRD §59, §60).
type ResolvedCapability struct {
	ID                 string
	Name               string
	ProviderType       entities.ProviderType
	ProviderID         string
	InputSchema        []byte
	OutputSchema       []byte
	CredentialReference string
}

// CapabilityResolver resolves a logical capability to its provider, schema, and
// effective permission honoring binding scope with most-restrictive-wins (PRD §60).
type CapabilityResolver struct {
	repo repositories.ICapabilityRepository
}

// NewCapabilityResolver builds a resolver backed by the capability repository.
func NewCapabilityResolver(repo repositories.ICapabilityRepository) *CapabilityResolver {
	return &CapabilityResolver{repo: repo}
}

// Resolve resolves a capability by name within a tenant and computes whether it
// is allowed for the given workflow/state. It returns the resolved capability
// and an error if the capability is unknown or denied. It never invokes a
// provider (PRD §60, §62).
func (r *CapabilityResolver) Resolve(ctx context.Context, tenantID, name, workflowID, stateID string) (*ResolvedCapability, error) {
	cap, err := r.repo.FindByName(ctx, tenantID, name)
	if err != nil {
		var de *domain.DomainError
		if errors.As(err, &de) && de.Code == domain.ErrNotFound {
			return nil, NewCapabilityError(ErrorKindUnauthorized, "capability.unauthorized", "capability not found: "+name)
		}
		return nil, err
	}
	if cap.Status != entities.CapabilityActive {
		return nil, NewCapabilityError(ErrorKindUnauthorized, "capability.unauthorized", "capability is not active: "+name)
	}

	// Binding resolution: Global → Tenant → Workflow → State (PRD §60).
	// DENY at a more specific level wins over ALLOW at a less specific level.
	permission, err := r.resolvePermission(ctx, tenantID, cap.ID, workflowID, stateID)
	if err != nil {
		return nil, err
	}
	if permission == entities.BindingPermissionDeny {
		return nil, NewCapabilityError(ErrorKindUnauthorized, "capability.unauthorized", "capability denied for context: "+name)
	}

	return &ResolvedCapability{
		ID:                  cap.ID,
		Name:                cap.Name,
		ProviderType:        cap.ProviderType,
		ProviderID:          cap.ProviderID.String,
		InputSchema:         cap.InputSchema,
		OutputSchema:        cap.OutputSchema,
		CredentialReference: cap.CredentialReference.String,
	}, nil
}

// resolvePermission walks bindings from least to most specific scope and
// returns the effective permission. Most-restrictive-wins: a DENY at a more
// specific scope (state > workflow > tenant) overrides any ALLOW.
func (r *CapabilityResolver) resolvePermission(ctx context.Context, tenantID, capabilityID, workflowID, stateID string) (entities.BindingPermission, error) {
	bindings, err := r.repo.ListBindingsByCapability(ctx, tenantID, capabilityID)
	if err != nil {
		return "", err
	}

	// most-restrictive-wins, most specific first
	scopes := []struct {
		typ entities.BindingScopeType
		id  string
	}{
		{entities.BindingScopeState, stateID},
		{entities.BindingScopeWorkflow, workflowID},
		{entities.BindingScopeTenant, tenantID},
	}
	for _, s := range scopes {
		for _, b := range bindings {
			if b.ScopeType == s.typ && (b.ScopeID == s.id || s.typ == entities.BindingScopeTenant) {
				if b.Permission == entities.BindingPermissionDeny {
					return entities.BindingPermissionDeny, nil
				}
			}
		}
	}
	// no DENY found → allowed by default (absent binding falls through)
	return entities.BindingPermissionAllow, nil
}
