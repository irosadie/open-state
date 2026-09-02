package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"reflect"
	"time"

	"github.com/google/uuid"
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
		Status: string(input.Status), Column13: nullableString(input.OAuthAuthorizationEndpoint),
		Column14: nullableString(input.OAuthTokenEndpoint), Column15: nullableString(input.OAuthClientID), Column16: nullableString(input.OAuthClientSecretReference),
		OauthScopes: marshalOAuthScopes(input.OAuthScopes), Column18: nullableString(input.OAuthRedirectURI),
		TimeoutMs: int32(input.TimeoutMS), MaxConcurrency: int32(input.MaxConcurrency),
		RateLimitPerSecond: input.RateLimitPerSecond, RateLimitBurst: int32(input.RateLimitBurst),
		RetryMax: int32(input.RetryMax), CircuitFailureThreshold: int32(input.CircuitFailureThreshold),
		CircuitRecoverySeconds: int32(input.CircuitRecoverySeconds), CreatedBy: input.CreatedBy,
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
		Column13: nullableString(input.OAuthAuthorizationEndpoint), Column14: nullableString(input.OAuthTokenEndpoint),
		Column15: nullableString(input.OAuthClientID), Column16: nullableString(input.OAuthClientSecretReference), OauthScopes: marshalOAuthScopes(input.OAuthScopes),
		Column18: nullableString(input.OAuthRedirectURI), TimeoutMs: int32(input.TimeoutMS),
		MaxConcurrency: int32(input.MaxConcurrency), RateLimitPerSecond: input.RateLimitPerSecond,
		RateLimitBurst: int32(input.RateLimitBurst), RetryMax: int32(input.RetryMax),
		CircuitFailureThreshold: int32(input.CircuitFailureThreshold), CircuitRecoverySeconds: int32(input.CircuitRecoverySeconds),
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

func (r *PgxMCPConnectionRepository) UpdateOAuth(ctx context.Context, input repositories.MCPConnectionOAuthUpdateInput) (*entities.MCPConnection, error) {
	row, err := r.queries.UpdateMCPConnectionOAuth(ctx, db.UpdateMCPConnectionOAuthParams{
		ID: mustUUID(input.ID), TenantID: mustUUID(input.TenantID), ProjectID: mustUUID(input.ProjectID),
		CredentialReference:        nullableStringValuePtr(input.CredentialReference),
		OauthRefreshTokenReference: nullableStringValuePtr(input.RefreshTokenReference),
		CredentialStatus:           string(input.CredentialStatus), OauthStatus: string(input.OAuthStatus),
		OauthExpiresAt: nullableTimeValue(input.OAuthExpiresAt), UpdatedBy: input.UpdatedBy,
	})
	if err != nil {
		return nil, mapNotFound(err, "MCP OAuth credentials")
	}
	return mapMCPConnection(row), nil
}

func (r *PgxMCPConnectionRepository) UpdateCredential(ctx context.Context, input repositories.MCPConnectionCredentialUpdateInput) (*entities.MCPConnection, error) {
	row, err := r.queries.UpdateMCPConnectionCredential(ctx, db.UpdateMCPConnectionCredentialParams{
		ID: mustUUID(input.ID), TenantID: mustUUID(input.TenantID), ProjectID: mustUUID(input.ProjectID),
		CredentialReference: nullableStringValuePtr(input.CredentialReference),
		CredentialStatus:    string(input.CredentialStatus), UpdatedBy: input.UpdatedBy,
	})
	if err != nil {
		return nil, mapNotFound(err, "MCP credential")
	}
	return mapMCPConnection(row), nil
}

func (r *PgxMCPConnectionRepository) DisconnectOAuth(ctx context.Context, tenantID, projectID, id, actor string) (*entities.MCPConnection, error) {
	row, err := r.queries.DisconnectMCPConnectionOAuth(ctx, db.DisconnectMCPConnectionOAuthParams{ID: mustUUID(id), TenantID: mustUUID(tenantID), ProjectID: mustUUID(projectID), UpdatedBy: actor})
	if err != nil {
		return nil, mapNotFound(err, "MCP OAuth connection")
	}
	return mapMCPConnection(row), nil
}

func (r *PgxMCPConnectionRepository) RecordHealth(ctx context.Context, input repositories.MCPConnectionHealthUpdateInput) (*entities.MCPConnection, error) {
	row, err := r.queries.RecordMCPConnectionHealth(ctx, db.RecordMCPConnectionHealthParams{
		ID: mustUUID(input.ID), TenantID: mustUUID(input.TenantID), ProjectID: mustUUID(input.ProjectID),
		HealthStatus: string(input.HealthStatus), Column5: nullableString(input.HealthReason),
		LastSuccessAt: nullableTimeValue(input.LastSuccessAt), ConsecutiveFailures: int32(input.ConsecutiveFailures),
		CircuitOpenedAt: nullableTimeValue(input.CircuitOpenedAt), UpdatedBy: input.Actor,
	})
	if err != nil {
		return nil, mapNotFound(err, "MCP connection")
	}
	return mapMCPConnection(row), nil
}

func (r *PgxMCPConnectionRepository) ResetHealth(ctx context.Context, tenantID, projectID, id, actor string) (*entities.MCPConnection, error) {
	row, err := r.queries.ResetMCPConnectionHealth(ctx, db.ResetMCPConnectionHealthParams{ID: mustUUID(id), TenantID: mustUUID(tenantID), ProjectID: mustUUID(projectID), UpdatedBy: actor})
	if err != nil {
		return nil, mapNotFound(err, "MCP connection")
	}
	return mapMCPConnection(row), nil
}

// sqlc emits a distinct row type for each RETURNING query. Reflection is kept
// inside this database adapter so the domain remains independent from those
// generated implementation details.
func mapMCPConnection(raw any) *entities.MCPConnection {
	row := reflect.ValueOf(raw)
	if row.Kind() == reflect.Pointer {
		row = row.Elem()
	}
	field := func(name string) reflect.Value { return row.FieldByName(name) }
	stringValue := func(name string) string { return reflectText(field(name)) }
	return &entities.MCPConnection{
		ID: stringValue("ID"), TenantID: stringValue("TenantID"), ProjectID: stringValue("ProjectID"),
		Name: stringValue("Name"), Alias: stringValue("Alias"), Transport: entities.MCPConnectionTransport(stringValue("Transport")),
		Endpoint: reflectStringPtr(field("Endpoint")), StdioProfile: reflectStringPtr(field("StdioProfile")),
		StdioArgs: unmarshalStdioArgs(field("StdioArgs").Interface().(json.RawMessage)),
		AuthType:  entities.MCPConnectionAuthType(stringValue("AuthType")), CredentialReference: reflectStringPtr(field("CredentialReference")),
		CredentialStatus: entities.MCPConnectionCredentialStatus(stringValue("CredentialStatus")), Status: entities.MCPConnectionStatus(stringValue("Status")),
		OAuthAuthorizationEndpoint: reflectStringPtr(field("OauthAuthorizationEndpoint")), OAuthTokenEndpoint: reflectStringPtr(field("OauthTokenEndpoint")),
		OAuthClientID: reflectStringPtr(field("OauthClientID")), OAuthClientSecretReference: reflectStringPtr(field("OauthClientSecretReference")), OAuthScopes: unmarshalOAuthScopes(field("OauthScopes").Interface().(json.RawMessage)),
		OAuthRedirectURI: reflectStringPtr(field("OauthRedirectUri")), OAuthAccessTokenReference: reflectStringPtr(field("OauthAccessTokenReference")),
		OAuthRefreshTokenReference: reflectStringPtr(field("OauthRefreshTokenReference")), OAuthExpiresAt: reflectTimePtr(field("OauthExpiresAt")),
		OAuthStatus: entities.MCPOAuthStatus(stringValue("OauthStatus")), HealthStatus: entities.MCPConnectionHealthStatus(stringValue("HealthStatus")),
		HealthReason: reflectStringPtr(field("HealthReason")), LastSuccessAt: reflectTimePtr(field("LastSuccessAt")),
		ConsecutiveFailures: int(field("ConsecutiveFailures").Int()), CircuitOpenedAt: reflectTimePtr(field("CircuitOpenedAt")),
		TimeoutMS: int(field("TimeoutMs").Int()), MaxConcurrency: int(field("MaxConcurrency").Int()), RateLimitPerSecond: field("RateLimitPerSecond").Float(),
		RateLimitBurst: int(field("RateLimitBurst").Int()), RetryMax: int(field("RetryMax").Int()),
		CircuitFailureThreshold: int(field("CircuitFailureThreshold").Int()), CircuitRecoverySeconds: int(field("CircuitRecoverySeconds").Int()),
		LastTestStatus: entities.MCPConnectionTestStatus(stringValue("LastTestStatus")), LastTestErrorCode: reflectStringPtr(field("LastTestErrorCode")),
		LastTestedAt: reflectTimePtr(field("LastTestedAt")), CreatedBy: stringValue("CreatedBy"), UpdatedBy: stringValue("UpdatedBy"),
		CreatedAt: field("CreatedAt").Interface().(time.Time), UpdatedAt: field("UpdatedAt").Interface().(time.Time),
	}
}

func reflectText(value reflect.Value) string {
	if !value.IsValid() {
		return ""
	}
	if id, ok := value.Interface().(uuid.UUID); ok {
		return id.String()
	}
	if value.Kind() == reflect.String {
		return value.String()
	}
	return ""
}

func reflectStringPtr(value reflect.Value) *string {
	if !value.IsValid() {
		return nil
	}
	if nullable, ok := value.Interface().(sql.NullString); ok && nullable.Valid {
		value := nullable.String
		return &value
	}
	return nil
}

func reflectTimePtr(value reflect.Value) *time.Time {
	if !value.IsValid() {
		return nil
	}
	if nullable, ok := value.Interface().(sql.NullTime); ok && nullable.Valid {
		value := nullable.Time
		return &value
	}
	return nil
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

func marshalOAuthScopes(scopes []string) json.RawMessage {
	if scopes == nil {
		scopes = []string{}
	}
	raw, _ := json.Marshal(scopes)
	return raw
}

func unmarshalOAuthScopes(raw json.RawMessage) []string {
	var scopes []string
	if err := json.Unmarshal(raw, &scopes); err != nil || scopes == nil {
		return []string{}
	}
	return scopes
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

func nullableTimeValue(value *time.Time) sql.NullTime {
	if value == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *value, Valid: true}
}

func mapWriteMCPError(err error, op string) error {
	if errors.Is(err, sql.ErrNoRows) {
		return domain.NewNotFound("MCP connection not found")
	}
	return mapPgError(err, op)
}
