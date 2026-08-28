package repositories

import (
	"context"

	"github.com/irosadie/open-state/api/internal/domain/entities"
)

// IUserIdentityRepository defines the persistence contract for external OIDC
// identities (PRD §79). Identities link a provider subject to a local user.
type IUserIdentityRepository interface {
	// FindByProviderSubject returns the identity for a provider subject, or nil.
	FindByProviderSubject(ctx context.Context, provider, subjectID string) (*entities.UserIdentity, error)
	// Create links an identity to a user (auto_provisioned marks first-login users).
	Create(ctx context.Context, userID, provider, subjectID string, autoProvisioned bool) (*entities.UserIdentity, error)
	// ListByUser returns all identities for a user.
	ListByUser(ctx context.Context, userID string) ([]entities.UserIdentity, error)
}
