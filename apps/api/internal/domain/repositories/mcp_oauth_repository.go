package repositories

import (
	"context"
	"time"

	"github.com/irosadie/open-state/api/internal/domain/entities"
)

type IMCPOAuthTransactionRepository interface {
	Create(ctx context.Context, input MCPAuthorizationTransactionCreateInput) (*entities.MCPAuthorizationTransaction, error)
	FindPendingByState(ctx context.Context, tenantID, projectID, connectionID string, stateHash []byte) (*entities.MCPAuthorizationTransaction, error)
	Consume(ctx context.Context, tenantID, projectID, id string) error
}

type MCPAuthorizationTransactionCreateInput struct {
	TenantID          string
	ProjectID         string
	ConnectionID      string
	StateHash         []byte
	VerifierReference string
	RedirectURI       string
	ExpiresAt         time.Time
}
