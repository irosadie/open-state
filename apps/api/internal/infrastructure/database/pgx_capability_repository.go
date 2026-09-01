package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	"github.com/irosadie/open-state/api/internal/infrastructure/db"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

// PgxCapabilityRepository implements ICapabilityRepository on PostgreSQL via sqlc.
type PgxCapabilityRepository struct {
	queries *db.Queries
}

// NewPgxCapabilityRepository returns a PostgreSQL-backed ICapabilityRepository.
func NewPgxCapabilityRepository(pool *pgxpool.Pool) repositories.ICapabilityRepository {
	sqlDB := stdlib.OpenDBFromPool(pool)
	return newPgxCapabilityRepository(db.New(sqlDB))
}

// newPgxCapabilityRepository builds an ICapabilityRepository from a sqlc queries
// handle. It enables the composed PostgresAdapter to bind this repository to a
// shared transaction via WithTx.
func newPgxCapabilityRepository(q *db.Queries) repositories.ICapabilityRepository {
	return &PgxCapabilityRepository{queries: q}
}

func (r *PgxCapabilityRepository) Create(ctx context.Context, tenantID, name string, description *string, providerType entities.ProviderType, providerID, providerTool *string, inputSchema, outputSchema []byte, version int, credentialReference *string) (*entities.Capability, error) {
	row, err := r.queries.CreateCapability(ctx, db.CreateCapabilityParams{
		TenantID:            mustUUID(tenantID),
		Name:                name,
		Description:         nullString(description),
		ProviderType:        string(providerType),
		ProviderID:          nullString(providerID),
		ProviderTool:        nullString(providerTool),
		InputSchema:         inputSchema,
		OutputSchema:        outputSchema,
		Status:              string(entities.CapabilityActive),
		Version:             int32(version),
		CredentialReference: nullString(credentialReference),
	})
	if err != nil {
		return nil, mapPgError(err, "create capability")
	}
	return mapCapability(row), nil
}

func (r *PgxCapabilityRepository) FindByID(ctx context.Context, tenantID, id string) (*entities.Capability, error) {
	row, err := r.queries.FindCapabilityByID(ctx, db.FindCapabilityByIDParams{
		ID:       mustUUID(id),
		TenantID: mustUUID(tenantID),
	})
	if err != nil {
		return nil, mapNotFound(err, "capability")
	}
	return mapCapability(row), nil
}

func (r *PgxCapabilityRepository) FindByName(ctx context.Context, tenantID, name string) (*entities.Capability, error) {
	row, err := r.queries.FindCapabilityByName(ctx, db.FindCapabilityByNameParams{
		Name:     name,
		TenantID: mustUUID(tenantID),
	})
	if err != nil {
		return nil, mapNotFound(err, "capability")
	}
	return mapCapability(row), nil
}

func (r *PgxCapabilityRepository) ListByTenant(ctx context.Context, tenantID string) ([]entities.Capability, error) {
	rows, err := r.queries.ListCapabilitiesByTenant(ctx, mustUUID(tenantID))
	if err != nil {
		return nil, err
	}
	out := make([]entities.Capability, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapCapability(row))
	}
	return out, nil
}

func (r *PgxCapabilityRepository) ListByTenantFiltered(ctx context.Context, tenantID string, providerType entities.ProviderType, capStatus entities.CapabilityStatus) ([]entities.Capability, error) {
	rows, err := r.queries.ListCapabilitiesByTenantFiltered(ctx, db.ListCapabilitiesByTenantFilteredParams{
		TenantID: mustUUID(tenantID),
		Column2:  string(providerType),
		Column3:  string(capStatus),
	})
	if err != nil {
		return nil, err
	}
	out := make([]entities.Capability, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapCapability(row))
	}
	return out, nil
}

func (r *PgxCapabilityRepository) Update(ctx context.Context, tenantID, id string, description *string, providerType entities.ProviderType, providerID, providerTool *string, inputSchema, outputSchema []byte, status entities.CapabilityStatus, version int, credentialReference *string) (*entities.Capability, error) {
	row, err := r.queries.UpdateCapability(ctx, db.UpdateCapabilityParams{
		ID:                  mustUUID(id),
		TenantID:            mustUUID(tenantID),
		Description:         nullString(description),
		ProviderType:        string(providerType),
		ProviderID:          nullString(providerID),
		ProviderTool:        nullString(providerTool),
		InputSchema:         inputSchema,
		OutputSchema:        outputSchema,
		Status:              string(status),
		Version:             int32(version),
		CredentialReference: nullString(credentialReference),
	})
	if err != nil {
		return nil, mapNotFound(err, "capability")
	}
	return mapCapability(row), nil
}

func (r *PgxCapabilityRepository) Disable(ctx context.Context, tenantID, id string) (*entities.Capability, error) {
	row, err := r.queries.DisableCapability(ctx, db.DisableCapabilityParams{
		ID:       mustUUID(id),
		TenantID: mustUUID(tenantID),
	})
	if err != nil {
		return nil, mapNotFound(err, "capability")
	}
	return mapCapability(row), nil
}

func (r *PgxCapabilityRepository) UpdateStatus(ctx context.Context, tenantID, id string, status entities.CapabilityStatus) (*entities.Capability, error) {
	row, err := r.queries.UpdateCapabilityStatus(ctx, db.UpdateCapabilityStatusParams{
		ID:       mustUUID(id),
		TenantID: mustUUID(tenantID),
		Status:   string(status),
	})
	if err != nil {
		return nil, mapNotFound(err, "capability")
	}
	return mapCapability(row), nil
}

func (r *PgxCapabilityRepository) Bind(ctx context.Context, tenantID, capabilityID string, scopeType entities.BindingScopeType, scopeID string, permission entities.BindingPermission) (*entities.CapabilityBinding, error) {
	row, err := r.queries.BindCapability(ctx, db.BindCapabilityParams{
		TenantID:     mustUUID(tenantID),
		CapabilityID: mustUUID(capabilityID),
		ScopeType:    string(scopeType),
		ScopeID:      scopeID,
		Permission:   string(permission),
	})
	if err != nil {
		return nil, mapPgError(err, "bind capability")
	}
	return mapCapabilityBinding(row), nil
}

func (r *PgxCapabilityRepository) ListBindingsByCapability(ctx context.Context, tenantID, capabilityID string) ([]entities.CapabilityBinding, error) {
	rows, err := r.queries.ListBindingsByCapability(ctx, db.ListBindingsByCapabilityParams{
		CapabilityID: mustUUID(capabilityID),
		TenantID:     mustUUID(tenantID),
	})
	if err != nil {
		return nil, err
	}
	out := make([]entities.CapabilityBinding, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapCapabilityBinding(row))
	}
	return out, nil
}

func (r *PgxCapabilityRepository) ListBindingsByScope(ctx context.Context, tenantID string, scopeType entities.BindingScopeType, scopeID string) ([]entities.CapabilityBinding, error) {
	rows, err := r.queries.ListBindingsByScope(ctx, db.ListBindingsByScopeParams{
		ScopeType: string(scopeType),
		ScopeID:   scopeID,
		TenantID:  mustUUID(tenantID),
	})
	if err != nil {
		return nil, err
	}
	out := make([]entities.CapabilityBinding, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapCapabilityBinding(row))
	}
	return out, nil
}

func (r *PgxCapabilityRepository) Unbind(ctx context.Context, tenantID, bindingID string) error {
	_, err := r.queries.DeleteBinding(ctx, db.DeleteBindingParams{
		ID:       mustUUID(bindingID),
		TenantID: mustUUID(tenantID),
	})
	if err != nil {
		return mapNotFound(err, "binding")
	}
	return nil
}

func (r *PgxCapabilityRepository) UpsertPolicy(ctx context.Context, tenantID string, scopeType entities.PolicyScopeType, scopeID, policyType string, content []byte) (*entities.Policy, error) {
	row, err := r.queries.UpsertPolicy(ctx, db.UpsertPolicyParams{
		TenantID:  mustUUID(tenantID),
		ScopeType: string(scopeType),
		ScopeID:   scopeID,
		Type:      policyType,
		Content:   content,
	})
	if err != nil {
		return nil, mapPgError(err, "upsert policy")
	}
	return mapPolicy(row), nil
}

func (r *PgxCapabilityRepository) FindPolicyByType(ctx context.Context, tenantID string, scopeType entities.PolicyScopeType, scopeID, policyType string) (*entities.Policy, error) {
	row, err := r.queries.FindPolicyByType(ctx, db.FindPolicyByTypeParams{
		TenantID:  mustUUID(tenantID),
		ScopeType: string(scopeType),
		ScopeID:   scopeID,
		Type:      policyType,
	})
	if err != nil {
		return nil, mapNotFound(err, "policy")
	}
	return mapPolicy(row), nil
}

func (r *PgxCapabilityRepository) ListPoliciesByScope(ctx context.Context, tenantID string, scopeType entities.PolicyScopeType, scopeID string) ([]entities.Policy, error) {
	rows, err := r.queries.ListPoliciesByScope(ctx, db.ListPoliciesByScopeParams{
		TenantID:  mustUUID(tenantID),
		ScopeType: string(scopeType),
		ScopeID:   scopeID,
	})
	if err != nil {
		return nil, err
	}
	out := make([]entities.Policy, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapPolicy(row))
	}
	return out, nil
}

// ---- mappers ----

func mapCapability(row any) *entities.Capability {
	switch value := row.(type) {
	case db.Capability:
		return mapCapabilityValues(value.ID.String(), value.TenantID.String(), value.Name, value.Description, value.ProviderType, value.ProviderID, value.ProviderTool, value.InputSchema, value.OutputSchema, value.Status, value.Version, value.CredentialReference, value.CreatedAt, value.UpdatedAt)
	case db.CreateCapabilityRow:
		return mapCapabilityValues(value.ID.String(), value.TenantID.String(), value.Name, value.Description, value.ProviderType, value.ProviderID, value.ProviderTool, value.InputSchema, value.OutputSchema, value.Status, value.Version, value.CredentialReference, value.CreatedAt, value.UpdatedAt)
	case db.FindCapabilityByIDRow:
		return mapCapabilityValues(value.ID.String(), value.TenantID.String(), value.Name, value.Description, value.ProviderType, value.ProviderID, value.ProviderTool, value.InputSchema, value.OutputSchema, value.Status, value.Version, value.CredentialReference, value.CreatedAt, value.UpdatedAt)
	case db.FindCapabilityByNameRow:
		return mapCapabilityValues(value.ID.String(), value.TenantID.String(), value.Name, value.Description, value.ProviderType, value.ProviderID, value.ProviderTool, value.InputSchema, value.OutputSchema, value.Status, value.Version, value.CredentialReference, value.CreatedAt, value.UpdatedAt)
	case db.ListCapabilitiesByTenantRow:
		return mapCapabilityValues(value.ID.String(), value.TenantID.String(), value.Name, value.Description, value.ProviderType, value.ProviderID, value.ProviderTool, value.InputSchema, value.OutputSchema, value.Status, value.Version, value.CredentialReference, value.CreatedAt, value.UpdatedAt)
	case db.ListCapabilitiesByTenantFilteredRow:
		return mapCapabilityValues(value.ID.String(), value.TenantID.String(), value.Name, value.Description, value.ProviderType, value.ProviderID, value.ProviderTool, value.InputSchema, value.OutputSchema, value.Status, value.Version, value.CredentialReference, value.CreatedAt, value.UpdatedAt)
	case db.UpdateCapabilityRow:
		return mapCapabilityValues(value.ID.String(), value.TenantID.String(), value.Name, value.Description, value.ProviderType, value.ProviderID, value.ProviderTool, value.InputSchema, value.OutputSchema, value.Status, value.Version, value.CredentialReference, value.CreatedAt, value.UpdatedAt)
	case db.DisableCapabilityRow:
		return mapCapabilityValues(value.ID.String(), value.TenantID.String(), value.Name, value.Description, value.ProviderType, value.ProviderID, value.ProviderTool, value.InputSchema, value.OutputSchema, value.Status, value.Version, value.CredentialReference, value.CreatedAt, value.UpdatedAt)
	case db.UpdateCapabilityStatusRow:
		return mapCapabilityValues(value.ID.String(), value.TenantID.String(), value.Name, value.Description, value.ProviderType, value.ProviderID, value.ProviderTool, value.InputSchema, value.OutputSchema, value.Status, value.Version, value.CredentialReference, value.CreatedAt, value.UpdatedAt)
	default:
		panic("unsupported sqlc capability row")
	}
}

func mapCapabilityValues(id, tenantID, name string, description sql.NullString, providerType string, providerID, providerTool sql.NullString, inputSchema, outputSchema []byte, status string, version int32, credentialReference sql.NullString, createdAt, updatedAt time.Time) *entities.Capability {
	return &entities.Capability{ID: id, TenantID: tenantID, Name: name, Description: description, ProviderType: entities.ProviderType(providerType), ProviderID: providerID, ProviderTool: providerTool, InputSchema: inputSchema, OutputSchema: outputSchema, Status: entities.CapabilityStatus(status), Version: int(version), CredentialReference: credentialReference, CreatedAt: createdAt, UpdatedAt: updatedAt}
}

func mapCapabilityBinding(row db.CapabilityBinding) *entities.CapabilityBinding {
	return &entities.CapabilityBinding{
		ID:           row.ID.String(),
		TenantID:     row.TenantID.String(),
		CapabilityID: row.CapabilityID.String(),
		ScopeType:    entities.BindingScopeType(row.ScopeType),
		ScopeID:      row.ScopeID,
		Permission:   entities.BindingPermission(row.Permission),
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}

func mapPolicy(row db.Policy) *entities.Policy {
	return &entities.Policy{
		ID:        row.ID.String(),
		TenantID:  row.TenantID.String(),
		ScopeType: entities.PolicyScopeType(row.ScopeType),
		ScopeID:   row.ScopeID,
		Type:      row.Type,
		Content:   row.Content,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}
