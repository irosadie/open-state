package database

import (
	"context"
	"database/sql"
	"errors"

	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	"github.com/irosadie/open-state/api/internal/infrastructure/db"
	domain "github.com/irosadie/open-state/go-shared/domain"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

type PgxMCPAuthorizationTransactionRepository struct{ queries *db.Queries }

func NewPgxMCPAuthorizationTransactionRepository(pool *pgxpool.Pool) repositories.IMCPOAuthTransactionRepository {
	return &PgxMCPAuthorizationTransactionRepository{queries: db.New(stdlib.OpenDBFromPool(pool))}
}

func newPgxMCPAuthorizationTransactionRepository(q *db.Queries) repositories.IMCPOAuthTransactionRepository {
	return &PgxMCPAuthorizationTransactionRepository{queries: q}
}

func (r *PgxMCPAuthorizationTransactionRepository) Create(ctx context.Context, input repositories.MCPAuthorizationTransactionCreateInput) (*entities.MCPAuthorizationTransaction, error) {
	row, err := r.queries.CreateMCPAuthorizationTransaction(ctx, db.CreateMCPAuthorizationTransactionParams{
		TenantID: mustUUID(input.TenantID), ProjectID: mustUUID(input.ProjectID), ConnectionID: mustUUID(input.ConnectionID),
		StateHash: input.StateHash, VerifierReference: input.VerifierReference, RedirectUri: input.RedirectURI, ExpiresAt: input.ExpiresAt,
	})
	if err != nil {
		return nil, mapPgError(err, "create MCP OAuth transaction")
	}
	return mapMCPAuthorizationTransaction(row), nil
}

func (r *PgxMCPAuthorizationTransactionRepository) FindPendingByState(ctx context.Context, tenantID, projectID, connectionID string, stateHash []byte) (*entities.MCPAuthorizationTransaction, error) {
	row, err := r.queries.FindPendingMCPAuthorizationTransaction(ctx, db.FindPendingMCPAuthorizationTransactionParams{TenantID: mustUUID(tenantID), ProjectID: mustUUID(projectID), ConnectionID: mustUUID(connectionID), StateHash: stateHash})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.NewNotFound("MCP OAuth transaction not found")
		}
		return nil, mapPgError(err, "find MCP OAuth transaction")
	}
	return mapMCPAuthorizationTransaction(row), nil
}

func (r *PgxMCPAuthorizationTransactionRepository) Consume(ctx context.Context, tenantID, projectID, id string) error {
	rows, err := r.queries.ConsumeMCPAuthorizationTransaction(ctx, db.ConsumeMCPAuthorizationTransactionParams{ID: mustUUID(id), TenantID: mustUUID(tenantID), ProjectID: mustUUID(projectID)})
	if err != nil {
		return mapPgError(err, "consume MCP OAuth transaction")
	}
	if rows != 1 {
		return domain.NewConflict("MCP OAuth transaction is no longer available")
	}
	return nil
}

func mapMCPAuthorizationTransaction(row db.McpOauthTransaction) *entities.MCPAuthorizationTransaction {
	return &entities.MCPAuthorizationTransaction{
		ID: row.ID.String(), TenantID: row.TenantID.String(), ProjectID: row.ProjectID.String(), ConnectionID: row.ConnectionID.String(),
		StateHash: append([]byte(nil), row.StateHash...), VerifierReference: row.VerifierReference, RedirectURI: row.RedirectUri,
		ExpiresAt: row.ExpiresAt, Status: entities.MCPAuthorizationTransactionStatus(row.Status), CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}
