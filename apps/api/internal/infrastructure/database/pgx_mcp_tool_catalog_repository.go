package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	"github.com/irosadie/open-state/api/internal/infrastructure/db"
)

// PgxMCPToolCatalogRepository persists sanitized, project-scoped MCP tool
// snapshots. Reconcile uses one SQL transaction so a successful refresh can
// never expose a half-written catalog.
type PgxMCPToolCatalogRepository struct {
	queries *db.Queries
	db      *sql.DB
}

func NewPgxMCPToolCatalogRepository(pool *pgxpool.Pool) repositories.IMCPToolCatalogRepository {
	sqlDB := stdlib.OpenDBFromPool(pool)
	return newPgxMCPToolCatalogRepository(db.New(sqlDB), sqlDB)
}

func newPgxMCPToolCatalogRepository(q *db.Queries, sqlDB *sql.DB) repositories.IMCPToolCatalogRepository {
	return &PgxMCPToolCatalogRepository{queries: q, db: sqlDB}
}

func (r *PgxMCPToolCatalogRepository) Get(ctx context.Context, tenantID, projectID, connectionID string) (*entities.MCPToolCatalog, error) {
	tools, err := r.queries.ListMCPDiscoveredTools(ctx, db.ListMCPDiscoveredToolsParams{
		TenantID: mustUUID(tenantID), ProjectID: mustUUID(projectID), ConnectionID: mustUUID(connectionID),
	})
	if err != nil {
		return nil, mapPgError(err, "get MCP tool catalog")
	}
	catalog := &entities.MCPToolCatalog{ConnectionID: connectionID, Tools: make([]entities.MCPDiscoveredTool, 0, len(tools))}
	for _, row := range tools {
		catalog.Tools = append(catalog.Tools, *mapMCPDiscoveredTool(row))
	}
	latest, err := r.queries.GetLatestMCPDiscoveryRun(ctx, db.GetLatestMCPDiscoveryRunParams{
		TenantID: mustUUID(tenantID), ProjectID: mustUUID(projectID), ConnectionID: mustUUID(connectionID),
	})
	if err == nil {
		catalog.LatestRun = mapMCPDiscoveryRun(latest)
	} else if err != sql.ErrNoRows {
		return nil, mapPgError(err, "get latest MCP discovery run")
	}
	lastSuccessful, err := r.queries.GetLastSuccessfulMCPDiscoveryRun(ctx, db.GetLastSuccessfulMCPDiscoveryRunParams{
		TenantID: mustUUID(tenantID), ProjectID: mustUUID(projectID), ConnectionID: mustUUID(connectionID),
	})
	if err == nil {
		catalog.LastSuccessfulRun = mapMCPDiscoveryRun(lastSuccessful)
	} else if err != sql.ErrNoRows {
		return nil, mapPgError(err, "get last successful MCP discovery run")
	}
	return catalog, nil
}

func (r *PgxMCPToolCatalogRepository) Reconcile(ctx context.Context, input repositories.MCPToolCatalogReconcileInput) (*entities.MCPDiscoveryRun, error) {
	if r.db == nil {
		return nil, mapPgError(sql.ErrConnDone, "reconcile MCP tool catalog")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, mapPgError(err, "start MCP tool catalog transaction")
	}
	defer tx.Rollback() //nolint:errcheck // rollback is a no-op after commit
	q := r.queries.WithTx(tx)
	now := time.Now().UTC()
	run, err := q.CreateMCPDiscoveryRun(ctx, db.CreateMCPDiscoveryRunParams{
		TenantID: mustUUID(input.TenantID), ProjectID: mustUUID(input.ProjectID), ConnectionID: mustUUID(input.ConnectionID),
		Status: string(entities.MCPDiscoverySucceeded), ToolCount: int32(len(input.Tools)),
		Column6: input.CatalogFingerprint, Column7: "", StartedAt: now, CompletedAt: now, CreatedBy: input.Actor,
	})
	if err != nil {
		return nil, mapPgError(err, "create MCP discovery run")
	}
	if err := q.MarkMCPDiscoveredToolsRemoved(ctx, db.MarkMCPDiscoveredToolsRemovedParams{
		TenantID: mustUUID(input.TenantID), ProjectID: mustUUID(input.ProjectID), ConnectionID: mustUUID(input.ConnectionID),
	}); err != nil {
		return nil, mapPgError(err, "mark removed MCP tools")
	}
	for _, tool := range input.Tools {
		if _, err := q.UpsertMCPDiscoveredTool(ctx, db.UpsertMCPDiscoveredToolParams{
			TenantID: mustUUID(input.TenantID), ProjectID: mustUUID(input.ProjectID), ConnectionID: mustUUID(input.ConnectionID),
			ToolName: tool.Name, Column5: tool.Title, Description: tool.Description,
			InputSchema: json.RawMessage(tool.InputSchema), Annotations: json.RawMessage(tool.Annotations),
			Fingerprint: tool.Fingerprint, DiscoveryRunID: uuid.NullUUID{UUID: run.ID, Valid: true},
		}); err != nil {
			return nil, mapPgError(err, "upsert MCP discovered tool")
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, mapPgError(err, "commit MCP tool catalog")
	}
	return mapMCPDiscoveryRun(run), nil
}

func (r *PgxMCPToolCatalogRepository) RecordFailure(ctx context.Context, input repositories.MCPToolCatalogFailureInput) (*entities.MCPDiscoveryRun, error) {
	now := time.Now().UTC()
	run, err := r.queries.CreateMCPDiscoveryRun(ctx, db.CreateMCPDiscoveryRunParams{
		TenantID: mustUUID(input.TenantID), ProjectID: mustUUID(input.ProjectID), ConnectionID: mustUUID(input.ConnectionID),
		Status: string(entities.MCPDiscoveryFailed), ToolCount: 0, Column6: "", Column7: input.ErrorCode,
		StartedAt: now, CompletedAt: now, CreatedBy: input.Actor,
	})
	if err != nil {
		return nil, mapPgError(err, "record MCP discovery failure")
	}
	return mapMCPDiscoveryRun(run), nil
}

func (r *PgxMCPToolCatalogRepository) SetEnabled(ctx context.Context, tenantID, projectID, connectionID, toolName string, enabled bool) (*entities.MCPDiscoveredTool, error) {
	row, err := r.queries.SetMCPDiscoveredToolEnabled(ctx, db.SetMCPDiscoveredToolEnabledParams{
		TenantID: mustUUID(tenantID), ProjectID: mustUUID(projectID), ConnectionID: mustUUID(connectionID),
		Enabled: enabled, ToolName: toolName,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, mapNotFound(err, "MCP discovered tool")
		}
		return nil, mapPgError(err, "set MCP tool enabled state")
	}
	return mapMCPDiscoveredTool(row), nil
}

func mapMCPDiscoveryRun(row db.McpDiscoveryRun) *entities.MCPDiscoveryRun {
	return &entities.MCPDiscoveryRun{
		ID: row.ID.String(), TenantID: row.TenantID.String(), ProjectID: row.ProjectID.String(), ConnectionID: row.ConnectionID.String(),
		Status: entities.MCPDiscoveryRunStatus(row.Status), ToolCount: int(row.ToolCount),
		CatalogFingerprint: sqlStringPtr(row.CatalogFingerprint), ErrorCode: sqlStringPtr(row.ErrorCode),
		StartedAt: row.StartedAt, CompletedAt: row.CompletedAt, CreatedBy: row.CreatedBy,
	}
}

func mapMCPDiscoveredTool(row db.McpDiscoveredTool) *entities.MCPDiscoveredTool {
	return &entities.MCPDiscoveredTool{
		ID: row.ID.String(), TenantID: row.TenantID.String(), ProjectID: row.ProjectID.String(), ConnectionID: row.ConnectionID.String(),
		Name: row.ToolName, Title: sqlStringPtr(row.Title), Description: row.Description,
		InputSchema: append(json.RawMessage(nil), row.InputSchema...), Annotations: append(json.RawMessage(nil), row.Annotations...),
		Fingerprint: row.Fingerprint, Enabled: row.Enabled,
		Availability: entities.MCPDiscoveredToolAvailability(row.Availability), DriftStatus: entities.MCPDiscoveredToolDriftStatus(row.DriftStatus),
		FirstSeenAt: row.FirstSeenAt, LastSeenAt: row.LastSeenAt, RemovedAt: sqlTimePtr(row.RemovedAt),
		DiscoveryRunID: nullableUUIDStringPtr(row.DiscoveryRunID), CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}

func nullableUUIDStringPtr(value uuid.NullUUID) *string {
	if !value.Valid {
		return nil
	}
	s := value.UUID.String()
	return &s
}
