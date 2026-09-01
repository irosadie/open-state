package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	"github.com/irosadie/open-state/api/internal/infrastructure/db"
)

// PgxAPIKeyRepository persists tenant-scoped State MCP machine credentials.
type PgxAPIKeyRepository struct {
	queries *db.Queries
}

// NewPgxAPIKeyRepository returns a PostgreSQL-backed API key repository.
func NewPgxAPIKeyRepository(pool *pgxpool.Pool) repositories.IAPIKeyRepository {
	sqlDB := stdlib.OpenDBFromPool(pool)
	return newPgxAPIKeyRepository(db.New(sqlDB))
}

func newPgxAPIKeyRepository(q *db.Queries) repositories.IAPIKeyRepository {
	return &PgxAPIKeyRepository{queries: q}
}

func (r *PgxAPIKeyRepository) Create(ctx context.Context, input repositories.APIKeyCreateInput) (*entities.APIKey, error) {
	row, err := r.queries.CreateAPIKey(ctx, db.CreateAPIKeyParams{
		TenantID:         mustUUID(input.TenantID),
		Name:             input.Name,
		KeyPrefix:        input.Prefix,
		KeyVerifier:      input.KeyVerifier,
		DefaultProjectID: apiKeyNullUUID(input.DefaultProjectID),
		ExpiresAt:        apiKeyNullTime(input.ExpiresAt),
		CreatedBy:        input.CreatedBy,
	})
	if err != nil {
		return nil, mapPgError(err, "create API key")
	}
	for _, projectID := range input.ProjectIDs {
		if err := r.queries.AddAPIKeyProject(ctx, db.AddAPIKeyProjectParams{
			ApiKeyID: row.ID, ProjectID: mustUUID(projectID),
		}); err != nil {
			return nil, mapPgError(err, "add API key project")
		}
	}
	for _, scope := range input.Scopes {
		if err := r.queries.AddAPIKeyScope(ctx, db.AddAPIKeyScopeParams{
			ApiKeyID: row.ID, Scope: string(scope),
		}); err != nil {
			return nil, mapPgError(err, "add API key scope")
		}
	}
	return r.hydrate(ctx, row)
}

func (r *PgxAPIKeyRepository) FindByPrefix(ctx context.Context, prefix string) (*entities.APIKey, error) {
	row, err := r.queries.FindAPIKeyByPrefix(ctx, prefix)
	if err != nil {
		return nil, mapNotFound(err, "API key")
	}
	return r.hydrate(ctx, row)
}

func (r *PgxAPIKeyRepository) ListByTenant(ctx context.Context, tenantID string) ([]entities.APIKey, error) {
	rows, err := r.queries.ListAPIKeysByTenant(ctx, mustUUID(tenantID))
	if err != nil {
		return nil, mapPgError(err, "list API keys")
	}
	keys := make([]entities.APIKey, 0, len(rows))
	for _, row := range rows {
		key, err := r.hydrate(ctx, row)
		if err != nil {
			return nil, err
		}
		keys = append(keys, *key)
	}
	return keys, nil
}

func (r *PgxAPIKeyRepository) Revoke(ctx context.Context, tenantID, keyID string) (*entities.APIKey, error) {
	row, err := r.queries.RevokeAPIKey(ctx, db.RevokeAPIKeyParams{ID: mustUUID(keyID), TenantID: mustUUID(tenantID)})
	if err != nil {
		return nil, mapNotFound(err, "API key")
	}
	return r.hydrate(ctx, row)
}

func (r *PgxAPIKeyRepository) TouchLastUsed(ctx context.Context, keyID string) error {
	if err := r.queries.TouchAPIKeyLastUsed(ctx, mustUUID(keyID)); err != nil {
		return mapPgError(err, "touch API key")
	}
	return nil
}

func (r *PgxAPIKeyRepository) hydrate(ctx context.Context, row db.AuthApiKey) (*entities.APIKey, error) {
	projects, err := r.queries.ListAPIKeyProjects(ctx, row.ID)
	if err != nil {
		return nil, mapPgError(err, "list API key projects")
	}
	scopes, err := r.queries.ListAPIKeyScopes(ctx, row.ID)
	if err != nil {
		return nil, mapPgError(err, "list API key scopes")
	}
	projectIDs := make([]string, 0, len(projects))
	for _, projectID := range projects {
		projectIDs = append(projectIDs, projectID.String())
	}
	keyScopes := make([]entities.MCPAPIScope, 0, len(scopes))
	for _, scope := range scopes {
		keyScopes = append(keyScopes, entities.MCPAPIScope(scope))
	}
	return &entities.APIKey{
		ID:               row.ID.String(),
		TenantID:         row.TenantID.String(),
		Name:             row.Name,
		Prefix:           row.KeyPrefix,
		KeyVerifier:      append([]byte(nil), row.KeyVerifier...),
		ProjectIDs:       projectIDs,
		DefaultProjectID: apiKeyNullUUIDPtr(row.DefaultProjectID),
		Scopes:           keyScopes,
		ExpiresAt:        apiKeyNullTimePtr(row.ExpiresAt),
		RevokedAt:        apiKeyNullTimePtr(row.RevokedAt),
		LastUsedAt:       apiKeyNullTimePtr(row.LastUsedAt),
		CreatedBy:        row.CreatedBy,
		CreatedAt:        row.CreatedAt,
		UpdatedAt:        row.UpdatedAt,
	}, nil
}

func apiKeyNullUUID(id *string) uuid.NullUUID {
	if id == nil {
		return uuid.NullUUID{}
	}
	return uuid.NullUUID{UUID: mustUUID(*id), Valid: true}
}

func apiKeyNullUUIDPtr(id uuid.NullUUID) *string {
	if !id.Valid {
		return nil
	}
	value := id.UUID.String()
	return &value
}

func apiKeyNullTime(value *time.Time) sql.NullTime {
	if value == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *value, Valid: true}
}

func apiKeyNullTimePtr(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	copy := value.Time
	return &copy
}
