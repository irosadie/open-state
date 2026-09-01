package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	"github.com/irosadie/open-state/api/internal/infrastructure/db"
	domain "github.com/irosadie/open-state/go-shared/domain"
)

// PgxMCPConnectionRepository persists project-scoped MCP connections.
type PgxMCPConnectionRepository struct{ queries *db.Queries }

func NewPgxMCPConnectionRepository(pool *pgxpool.Pool) repositories.IMCPConnectionRepository {
	return newPgxMCPConnectionRepository(db.New(stdlib.OpenDBFromPool(pool)))
}

func newPgxMCPConnectionRepository(q *db.Queries) repositories.IMCPConnectionRepository {
	return &PgxMCPConnectionRepository{queries: q}
}

func (r *PgxMCPConnectionRepository) Create(ctx context.Context, input repositories.MCPConnectionCreateInput) (*entities.MCPConnection, error) {
	args, err := marshalStdioArgs(input.StdioArgs)
	if err != nil {
		return nil, err
	}
	row, err := r.queries.CreateMCPConnection(ctx, db.CreateMCPConnectionParams{
		TenantID: mustUUID(input.TenantID), ProjectID: mustUUID(input.ProjectID), Name: input.Name,
		Alias: input.Alias, Transport: string(input.Transport), Column6: nullableString(input.Endpoint),
		Column7: nullableString(input.StdioProfile), StdioArgs: args, AuthType: string(input.AuthType),
		Column10: nullableString(input.CredentialReference), CredentialStatus: string(input.CredentialStatus),
		Status: string(input.Status), CreatedBy: input.CreatedBy,
	})
	if err != nil {
		return nil, mapPgError(err, "create MCP connection")
	}
	return mapMCPConnection(row), nil
}

func (r *PgxMCPConnectionRepository) FindByID(ctx context.Context, tenantID, projectID, id string) (*entities.MCPConnection, error) {
	row, err := r.queries.FindMCPConnectionByID(ctx, db.FindMCPConnectionByIDParams{ID: mustUUID(id), TenantID: mustUUID(tenantID), ProjectID: mustUUID(projectID)})
	if err != nil {
		return nil, mapNotFound(err, "MCP connection")
	}
	return mapMCPConnection(row), nil
}

func (r *PgxMCPConnectionRepository) ListByProject(ctx context.Context, tenantID, projectID string) ([]entities.MCPConnection, error) {
	rows, err := r.queries.ListMCPConnectionsByProject(ctx, db.ListMCPConnectionsByProjectParams{TenantID: mustUUID(tenantID), ProjectID: mustUUID(projectID)})
	if err != nil {
		return nil, mapPgError(err, "list MCP connections")
	}
	out := make([]entities.MCPConnection, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapMCPConnection(row))
	}
	return out, nil
}

func (r *PgxMCPConnectionRepository) Update(ctx context.Context, input repositories.MCPConnectionUpdateInput) (*entities.MCPConnection, error) {
	args, err := marshalStdioArgs(input.StdioArgs)
	if err != nil {
		return nil, err
	}
	row, err := r.queries.UpdateMCPConnection(ctx, db.UpdateMCPConnectionParams{
		ID: mustUUID(input.ID), TenantID: mustUUID(input.TenantID), ProjectID: mustUUID(input.ProjectID),
		Name: input.Name, Alias: input.Alias, Transport: string(input.Transport), Column7: nullableString(input.Endpoint),
		Column8: nullableString(input.StdioProfile), StdioArgs: args, AuthType: string(input.AuthType),
		CredentialReference: nullableStringValuePtr(input.CredentialReference), CredentialStatus: string(input.CredentialStatus), UpdatedBy: input.UpdatedBy,
	})
	if err != nil {
		return nil, mapWriteMCPError(err, "update MCP connection")
	}
	return mapMCPConnection(row), nil
}

func (r *PgxMCPConnectionRepository) Delete(ctx context.Context, tenantID, projectID, id string) error {
	err := r.queries.DeleteMCPConnection(ctx, db.DeleteMCPConnectionParams{ID: mustUUID(id), TenantID: mustUUID(tenantID), ProjectID: mustUUID(projectID)})
	if err != nil {
		return mapPgError(err, "delete MCP connection")
	}
	return nil
}

func (r *PgxMCPConnectionRepository) UpdateStatus(ctx context.Context, tenantID, projectID, id string, status entities.MCPConnectionStatus, actor string) (*entities.MCPConnection, error) {
	row, err := r.queries.UpdateMCPConnectionStatus(ctx, db.UpdateMCPConnectionStatusParams{ID: mustUUID(id), TenantID: mustUUID(tenantID), ProjectID: mustUUID(projectID), Status: string(status), UpdatedBy: actor})
	if err != nil {
		return nil, mapNotFound(err, "MCP connection")
	}
	return mapMCPConnection(row), nil
}

func (r *PgxMCPConnectionRepository) RecordTest(ctx context.Context, tenantID, projectID, id string, status entities.MCPConnectionTestStatus, errorCode, actor string) (*entities.MCPConnection, error) {
	row, err := r.queries.RecordMCPConnectionTest(ctx, db.RecordMCPConnectionTestParams{ID: mustUUID(id), TenantID: mustUUID(tenantID), ProjectID: mustUUID(projectID), LastTestStatus: string(status), Column5: nullableStringValue(errorCode), UpdatedBy: actor})
	if err != nil {
		return nil, mapNotFound(err, "MCP connection")
	}
	return mapMCPConnection(row), nil
}

func mapMCPConnection(row db.McpConnection) *entities.MCPConnection {
	return &entities.MCPConnection{
		ID: row.ID.String(), TenantID: row.TenantID.String(), ProjectID: row.ProjectID.String(), Name: row.Name,
		Alias: row.Alias, Transport: entities.MCPConnectionTransport(row.Transport), Endpoint: sqlStringPtr(row.Endpoint),
		StdioProfile: sqlStringPtr(row.StdioProfile), StdioArgs: unmarshalStdioArgs(row.StdioArgs),
		AuthType: entities.MCPConnectionAuthType(row.AuthType), CredentialReference: sqlStringPtr(row.CredentialReference),
		CredentialStatus: entities.MCPConnectionCredentialStatus(row.CredentialStatus), Status: entities.MCPConnectionStatus(row.Status),
		LastTestStatus: entities.MCPConnectionTestStatus(row.LastTestStatus), LastTestErrorCode: sqlStringPtr(row.LastTestErrorCode),
		LastTestedAt: sqlTimePtr(row.LastTestedAt), CreatedBy: row.CreatedBy, UpdatedBy: row.UpdatedBy,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}

func marshalStdioArgs(args []string) (json.RawMessage, error) {
	if args == nil {
		args = []string{}
	}
	return json.Marshal(args)
}

func unmarshalStdioArgs(raw json.RawMessage) []string {
	var args []string
	if err := json.Unmarshal(raw, &args); err != nil || args == nil {
		return []string{}
	}
	return args
}

func nullableString(value *string) any {
	if value == nil {
		return ""
	}
	return *value
}

func nullableStringValue(value string) any {
	if value == "" {
		return ""
	}
	return value
}

func nullableStringValuePtr(value *string) sql.NullString {
	if value == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *value, Valid: true}
}

func sqlStringPtr(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	s := value.String
	return &s
}

func sqlTimePtr(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	t := value.Time
	return &t
}

func mapWriteMCPError(err error, op string) error {
	if errors.Is(err, sql.ErrNoRows) {
		return domain.NewNotFound("MCP connection not found")
	}
	return mapPgError(err, op)
}
