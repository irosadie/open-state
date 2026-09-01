package database

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	"github.com/irosadie/open-state/api/internal/infrastructure/db"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

// PgxProjectCapabilityMCPBindingRepository persists the explicit project
// capability-to-tool mapping and derives health from the live catalog rows.
type PgxProjectCapabilityMCPBindingRepository struct {
	queries *db.Queries
}

func NewPgxProjectCapabilityMCPBindingRepository(pool *pgxpool.Pool) repositories.IProjectCapabilityMCPBindingRepository {
	return newPgxProjectCapabilityMCPBindingRepository(db.New(stdlib.OpenDBFromPool(pool)))
}

func newPgxProjectCapabilityMCPBindingRepository(q *db.Queries) repositories.IProjectCapabilityMCPBindingRepository {
	return &PgxProjectCapabilityMCPBindingRepository{queries: q}
}

func (r *PgxProjectCapabilityMCPBindingRepository) ListEligibleToolOptions(ctx context.Context, tenantID, projectID string) ([]entities.ProjectMCPToolOption, error) {
	rows, err := r.queries.ListEligibleMCPToolOptions(ctx, db.ListEligibleMCPToolOptionsParams{
		TenantID: mustUUID(tenantID), ProjectID: mustUUID(projectID),
	})
	if err != nil {
		return nil, mapPgError(err, "list eligible MCP tool options")
	}
	options := make([]entities.ProjectMCPToolOption, 0, len(rows))
	for _, row := range rows {
		options = append(options, entities.ProjectMCPToolOption{
			ConnectionID: row.ConnectionID.String(), ConnectionName: row.ConnectionName,
			ConnectionAlias: row.ConnectionAlias, ConnectionStatus: entities.MCPConnectionStatus(row.ConnectionStatus),
			ToolID: row.ToolID.String(), ToolName: row.ToolName, ToolTitle: sqlStringPtr(row.ToolTitle),
			ToolDescription: row.ToolDescription, InputSchema: append(json.RawMessage(nil), row.ToolInputSchema...),
			ToolFingerprint: row.ToolFingerprint,
		})
	}
	return options, nil
}

func (r *PgxProjectCapabilityMCPBindingRepository) ListByProject(ctx context.Context, tenantID, projectID string) ([]entities.ProjectCapabilityMCPBinding, error) {
	rows, err := r.queries.ListProjectCapabilityMCPBindings(ctx, db.ListProjectCapabilityMCPBindingsParams{
		TenantID: mustUUID(tenantID), ProjectID: mustUUID(projectID),
	})
	if err != nil {
		return nil, mapPgError(err, "list project MCP bindings")
	}
	bindings := make([]entities.ProjectCapabilityMCPBinding, 0, len(rows))
	for _, row := range rows {
		bindings = append(bindings, mapProjectCapabilityMCPBinding(row))
	}
	return bindings, nil
}

func (r *PgxProjectCapabilityMCPBindingRepository) FindByCapability(ctx context.Context, tenantID, projectID, capabilityID string) (*entities.ProjectCapabilityMCPBinding, error) {
	row, err := r.queries.FindProjectCapabilityMCPBinding(ctx, db.FindProjectCapabilityMCPBindingParams{
		TenantID: mustUUID(tenantID), ProjectID: mustUUID(projectID), CapabilityID: mustUUID(capabilityID),
	})
	if err != nil {
		return nil, mapNotFound(err, "project MCP capability binding")
	}
	binding := mapFindProjectCapabilityMCPBinding(row)
	return &binding, nil
}

func (r *PgxProjectCapabilityMCPBindingRepository) Upsert(ctx context.Context, input repositories.ProjectCapabilityMCPBindingUpsertInput) error {
	err := r.queries.UpsertProjectCapabilityMCPBinding(ctx, db.UpsertProjectCapabilityMCPBindingParams{
		TenantID: mustUUID(input.TenantID), ProjectID: mustUUID(input.ProjectID), CapabilityID: mustUUID(input.CapabilityID),
		McpConnectionID: mustUUID(input.ConnectionID), McpDiscoveredToolID: mustUUID(input.ToolID),
		BoundToolFingerprint: input.ToolFingerprint,
	})
	if err != nil {
		return mapPgError(err, "upsert project MCP capability binding")
	}
	return nil
}

func (r *PgxProjectCapabilityMCPBindingRepository) Delete(ctx context.Context, tenantID, projectID, capabilityID string) error {
	if err := r.queries.DeleteProjectCapabilityMCPBinding(ctx, db.DeleteProjectCapabilityMCPBindingParams{
		TenantID: mustUUID(tenantID), ProjectID: mustUUID(projectID), CapabilityID: mustUUID(capabilityID),
	}); err != nil {
		if err == sql.ErrNoRows {
			return mapNotFound(err, "project MCP capability binding")
		}
		return mapPgError(err, "delete project MCP capability binding")
	}
	return nil
}

func mapProjectCapabilityMCPBinding(row db.ListProjectCapabilityMCPBindingsRow) entities.ProjectCapabilityMCPBinding {
	return entities.ProjectCapabilityMCPBinding{
		ID: row.ID.String(), TenantID: row.TenantID.String(), ProjectID: row.ProjectID.String(),
		CapabilityID: row.CapabilityID.String(), CapabilityName: row.CapabilityName,
		CapabilityDescription: sqlStringPtr(row.CapabilityDescription), ConnectionID: row.ConnectionID.String(),
		ConnectionName: row.ConnectionName, ConnectionAlias: row.ConnectionAlias,
		ConnectionStatus: entities.MCPConnectionStatus(row.ConnectionStatus), ToolID: row.ToolID.String(),
		ToolName: row.ToolName, ToolTitle: sqlStringPtr(row.ToolTitle), ToolDescription: row.ToolDescription,
		BoundToolFingerprint: row.BoundToolFingerprint, CurrentToolFingerprint: row.CurrentToolFingerprint,
		ToolEnabled: row.ToolEnabled, ToolAvailability: entities.MCPDiscoveredToolAvailability(row.ToolAvailability),
		ToolDriftStatus: entities.MCPDiscoveredToolDriftStatus(row.ToolDriftStatus),
		Health:          entities.ProjectCapabilityMCPBindingHealth(row.Health), HealthReason: row.HealthReason,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}

func mapFindProjectCapabilityMCPBinding(row db.FindProjectCapabilityMCPBindingRow) entities.ProjectCapabilityMCPBinding {
	return entities.ProjectCapabilityMCPBinding{
		ID: row.ID.String(), TenantID: row.TenantID.String(), ProjectID: row.ProjectID.String(),
		CapabilityID: row.CapabilityID.String(), CapabilityName: row.CapabilityName,
		CapabilityDescription: sqlStringPtr(row.CapabilityDescription), ConnectionID: row.ConnectionID.String(),
		ConnectionName: row.ConnectionName, ConnectionAlias: row.ConnectionAlias,
		ConnectionStatus: entities.MCPConnectionStatus(row.ConnectionStatus), ToolID: row.ToolID.String(),
		ToolName: row.ToolName, ToolTitle: sqlStringPtr(row.ToolTitle), ToolDescription: row.ToolDescription,
		BoundToolFingerprint: row.BoundToolFingerprint, CurrentToolFingerprint: row.CurrentToolFingerprint,
		ToolEnabled: row.ToolEnabled, ToolAvailability: entities.MCPDiscoveredToolAvailability(row.ToolAvailability),
		ToolDriftStatus: entities.MCPDiscoveredToolDriftStatus(row.ToolDriftStatus),
		Health:          entities.ProjectCapabilityMCPBindingHealth(row.Health), HealthReason: row.HealthReason,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}
