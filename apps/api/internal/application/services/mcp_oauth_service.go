package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/irosadie/open-state/api/internal/application/dtos"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	domainsvc "github.com/irosadie/open-state/api/internal/domain/services"
	domain "github.com/irosadie/open-state/go-shared/domain"
)

const oauthTransactionLifetime = 10 * time.Minute

type MCPOAuthService struct {
	connections  repositories.IMCPConnectionRepository
	transactions repositories.IMCPOAuthTransactionRepository
	secrets      domainsvc.SecretStore
	client       domainsvc.OAuthClient
	audit        *AuditWriter
}

func NewMCPOAuthService(connections repositories.IMCPConnectionRepository, transactions repositories.IMCPOAuthTransactionRepository, secrets domainsvc.SecretStore, client domainsvc.OAuthClient, audit *AuditWriter) *MCPOAuthService {
	return &MCPOAuthService{connections: connections, transactions: transactions, secrets: secrets, client: client, audit: audit}
}

func (s *MCPOAuthService) Start(ctx context.Context, tenantID, projectID, connectionID, actor string) (*dtos.MCPOAuthStartDTO, error) {
	connection, err := s.findOAuthConnection(ctx, tenantID, projectID, connectionID)
	if err != nil {
		return nil, err
	}
	if connection.AuthType != entities.MCPAuthOAuth {
		return nil, domain.NewValidation("connection does not use OAuth")
	}
	if s.secrets == nil || s.transactions == nil || s.client == nil {
		return nil, domain.NewInternal("OAuth lifecycle is not configured")
	}
	if connection.OAuthAuthorizationEndpoint == nil || connection.OAuthClientID == nil || connection.OAuthRedirectURI == nil {
		return nil, domain.NewConflict("OAuth connection requires authorization settings")
	}
	state, err := randomOAuthValue(32)
	if err != nil {
		return nil, domain.NewInternal("could not create OAuth transaction")
	}
	verifier, err := randomOAuthValue(48)
	if err != nil {
		return nil, domain.NewInternal("could not create OAuth transaction")
	}
	verifierRef, err := s.secrets.Put(ctx, tenantID, projectID, "mcp_oauth_pkce_verifier", verifier)
	if err != nil {
		return nil, domain.NewInternal("could not protect OAuth transaction")
	}
	expiresAt := time.Now().UTC().Add(oauthTransactionLifetime)
	_, err = s.transactions.Create(ctx, repositories.MCPAuthorizationTransactionCreateInput{
		TenantID: tenantID, ProjectID: projectID, ConnectionID: connectionID, StateHash: hashOAuthValue(state),
		VerifierReference: verifierRef, RedirectURI: *connection.OAuthRedirectURI, ExpiresAt: expiresAt,
	})
	if err != nil {
		_ = s.secrets.Revoke(ctx, tenantID, projectID, verifierRef)
		return nil, domain.NewInternal("could not create OAuth transaction")
	}
	challenge := base64.RawURLEncoding.EncodeToString(hashOAuthBytes([]byte(verifier)))
	authorization, err := url.Parse(*connection.OAuthAuthorizationEndpoint)
	if err != nil || authorization.Scheme != "https" || authorization.Host == "" || authorization.User != nil {
		_ = s.secrets.Revoke(ctx, tenantID, projectID, verifierRef)
		return nil, domain.NewConflict("OAuth authorization endpoint is invalid")
	}
	query := authorization.Query()
	query.Set("response_type", "code")
	query.Set("client_id", *connection.OAuthClientID)
	query.Set("redirect_uri", *connection.OAuthRedirectURI)
	query.Set("state", state)
	query.Set("code_challenge", challenge)
	query.Set("code_challenge_method", "S256")
	if len(connection.OAuthScopes) > 0 {
		query.Set("scope", strings.Join(connection.OAuthScopes, " "))
	}
	authorization.RawQuery = query.Encode()
	s.auditOAuth(ctx, tenantID, actor, connection, entities.AuditActionMCPOAuthStarted)
	return &dtos.MCPOAuthStartDTO{AuthorizationURL: authorization.String(), Status: string(entities.MCPOAuthActionRequired), ExpiresAt: expiresAt.Format(time.RFC3339)}, nil
}

func (s *MCPOAuthService) Callback(ctx context.Context, tenantID, projectID, connectionID, actor, state, code string) (*dtos.MCPConnectionDTO, error) {
	if strings.TrimSpace(state) == "" || strings.TrimSpace(code) == "" {
		return nil, domain.NewValidation("OAuth callback is missing state or code")
	}
	connection, err := s.findOAuthConnection(ctx, tenantID, projectID, connectionID)
	if err != nil {
		return nil, err
	}
	if s.transactions == nil || s.secrets == nil || s.client == nil || connection.OAuthTokenEndpoint == nil || connection.OAuthClientID == nil || connection.OAuthRedirectURI == nil {
		return nil, domain.NewInternal("OAuth lifecycle is not configured")
	}
	transaction, err := s.transactions.FindPendingByState(ctx, tenantID, projectID, connectionID, hashOAuthValue(state))
	if err != nil {
		return nil, domain.NewUnauthorized("OAuth callback state is invalid or expired")
	}
	verifier, err := s.secrets.Resolve(ctx, tenantID, projectID, transaction.VerifierReference)
	if err != nil {
		return nil, domain.NewUnauthorized("OAuth callback transaction is unavailable")
	}
	clientSecret := ""
	if connection.OAuthClientSecretReference != nil {
		clientSecret, err = s.secrets.Resolve(ctx, tenantID, projectID, *connection.OAuthClientSecretReference)
		if err != nil {
			return nil, domain.NewUnauthorized("OAuth client authorization is unavailable")
		}
	}
	// Claim the transaction before contacting the provider. The database update
	// is conditional on status=pending, so concurrent callbacks cannot redeem the
	// same authorization code twice.
	if err := s.transactions.Consume(ctx, tenantID, projectID, transaction.ID); err != nil {
		return nil, domain.NewUnauthorized("OAuth callback transaction is no longer available")
	}
	tokens, err := s.client.Exchange(ctx, *connection.OAuthTokenEndpoint, *connection.OAuthClientID, clientSecret, code, transaction.RedirectURI, verifier)
	if err != nil || strings.TrimSpace(tokens.AccessToken) == "" {
		return nil, domain.NewUnauthorized("OAuth provider authorization failed")
	}
	accessRef, err := s.secrets.Put(ctx, tenantID, projectID, "mcp_oauth_access_token", tokens.AccessToken)
	if err != nil {
		return nil, domain.NewInternal("OAuth access token could not be protected")
	}
	refreshRef := connection.OAuthRefreshTokenReference
	if tokens.RefreshToken != "" {
		ref, putErr := s.secrets.Put(ctx, tenantID, projectID, "mcp_oauth_refresh_token", tokens.RefreshToken)
		if putErr != nil {
			_ = s.secrets.Revoke(ctx, tenantID, projectID, accessRef)
			return nil, domain.NewInternal("OAuth refresh token could not be protected")
		}
		refreshRef = &ref
	}
	securityRepo, ok := s.connections.(repositories.IMCPConnectionSecurityRepository)
	if !ok {
		return nil, domain.NewInternal("MCP connection security repository is not configured")
	}
	expiresAt := time.Now().UTC().Add(tokens.ExpiresIn)
	updated, err := securityRepo.UpdateOAuth(ctx, repositories.MCPConnectionOAuthUpdateInput{TenantID: tenantID, ProjectID: projectID, ID: connectionID, CredentialReference: &accessRef, RefreshTokenReference: refreshRef, CredentialStatus: entities.MCPCredentialConfigured, OAuthStatus: entities.MCPOAuthConnected, OAuthExpiresAt: &expiresAt, UpdatedBy: actor})
	if err != nil {
		_ = s.secrets.Revoke(ctx, tenantID, projectID, accessRef)
		if refreshRef != nil && (connection.OAuthRefreshTokenReference == nil || *refreshRef != *connection.OAuthRefreshTokenReference) {
			_ = s.secrets.Revoke(ctx, tenantID, projectID, *refreshRef)
		}
		return nil, err
	}
	for _, ref := range []*string{connection.OAuthAccessTokenReference, connection.OAuthRefreshTokenReference} {
		if ref != nil && *ref != accessRef {
			_ = s.secrets.Revoke(ctx, tenantID, projectID, *ref)
		}
	}
	s.auditOAuth(ctx, tenantID, actor, updatedEntity(updated), entities.AuditActionMCPOAuthConnected)
	dto := toMCPConnectionDTO(updated)
	return &dto, nil
}

func (s *MCPOAuthService) Disconnect(ctx context.Context, tenantID, projectID, connectionID, actor string) (*dtos.MCPConnectionDTO, error) {
	connection, err := s.findOAuthConnection(ctx, tenantID, projectID, connectionID)
	if err != nil {
		return nil, err
	}
	if s.secrets == nil {
		return nil, domain.NewInternal("OAuth secret store is not configured")
	}
	if securityRepo, ok := s.connections.(repositories.IMCPConnectionSecurityRepository); ok {
		for _, ref := range []*string{connection.OAuthAccessTokenReference, connection.OAuthRefreshTokenReference} {
			if ref != nil {
				_ = s.secrets.Revoke(ctx, tenantID, projectID, *ref)
			}
		}
		updated, err := securityRepo.DisconnectOAuth(ctx, tenantID, projectID, connectionID, actor)
		if err != nil {
			return nil, err
		}
		s.auditOAuth(ctx, tenantID, actor, updatedEntity(updated), entities.AuditActionMCPOAuthDisconnected)
		dto := toMCPConnectionDTO(updated)
		return &dto, nil
	}
	return nil, domain.NewInternal("MCP connection security repository is not configured")
}

func (s *MCPOAuthService) Status(ctx context.Context, tenantID, projectID, connectionID string) (*dtos.MCPOAuthStatusDTO, error) {
	connection, err := s.findOAuthConnection(ctx, tenantID, projectID, connectionID)
	if err != nil {
		return nil, err
	}
	return &dtos.MCPOAuthStatusDTO{Status: string(connection.OAuthStatus), ExpiresAt: formatMCPTimePtr(connection.OAuthExpiresAt), CredentialStatus: string(connection.CredentialStatus)}, nil
}

// AccessToken implements the server-side refresh boundary used by the MCP
// transport. The raw token never crosses this service's HTTP/MCP interfaces.
func (s *MCPOAuthService) AccessToken(ctx context.Context, connectionID, tenantID, projectID string) (string, error) {
	connection, err := s.findOAuthConnection(ctx, tenantID, projectID, connectionID)
	if err != nil {
		return "", err
	}
	if s.secrets == nil || s.client == nil {
		return "", s.markActionRequired(ctx, connection, "oauth_lifecycle_unavailable")
	}
	if connection.OAuthAccessTokenReference != nil && (connection.OAuthExpiresAt == nil || time.Now().UTC().Before(connection.OAuthExpiresAt.Add(-30*time.Second))) {
		return s.secrets.Resolve(ctx, tenantID, projectID, *connection.OAuthAccessTokenReference)
	}
	if connection.OAuthRefreshTokenReference == nil || connection.OAuthTokenEndpoint == nil || connection.OAuthClientID == nil {
		return "", s.markActionRequired(ctx, connection, "oauth_token_expired")
	}
	refreshToken, err := s.secrets.Resolve(ctx, tenantID, projectID, *connection.OAuthRefreshTokenReference)
	if err != nil {
		return "", s.markActionRequired(ctx, connection, "oauth_refresh_credential_unavailable")
	}
	clientSecret := ""
	if connection.OAuthClientSecretReference != nil {
		clientSecret, err = s.secrets.Resolve(ctx, tenantID, projectID, *connection.OAuthClientSecretReference)
		if err != nil {
			return "", s.markActionRequired(ctx, connection, "oauth_client_credential_unavailable")
		}
	}
	tokens, err := s.client.Refresh(ctx, *connection.OAuthTokenEndpoint, *connection.OAuthClientID, clientSecret, refreshToken)
	if err != nil || tokens.AccessToken == "" {
		return "", s.markActionRequired(ctx, connection, "oauth_refresh_failed")
	}
	accessRef, err := s.secrets.Put(ctx, tenantID, projectID, "mcp_oauth_access_token", tokens.AccessToken)
	if err != nil {
		return "", s.markActionRequired(ctx, connection, "oauth_access_token_protection_failed")
	}
	refreshRef := connection.OAuthRefreshTokenReference
	if tokens.RefreshToken != "" {
		ref, putErr := s.secrets.Put(ctx, tenantID, projectID, "mcp_oauth_refresh_token", tokens.RefreshToken)
		if putErr != nil {
			return "", s.markActionRequired(ctx, connection, "oauth_refresh_token_protection_failed")
		}
		refreshRef = &ref
	}
	expiresAt := time.Now().UTC().Add(tokens.ExpiresIn)
	securityRepo, ok := s.connections.(repositories.IMCPConnectionSecurityRepository)
	if !ok {
		return "", errors.New("MCP connection security repository is not configured")
	}
	if _, err := securityRepo.UpdateOAuth(ctx, repositories.MCPConnectionOAuthUpdateInput{TenantID: tenantID, ProjectID: projectID, ID: connection.ID, CredentialReference: &accessRef, RefreshTokenReference: refreshRef, CredentialStatus: entities.MCPCredentialConfigured, OAuthStatus: entities.MCPOAuthConnected, OAuthExpiresAt: &expiresAt, UpdatedBy: "system"}); err != nil {
		return "", err
	}
	return tokens.AccessToken, nil
}

func (s *MCPOAuthService) markActionRequired(ctx context.Context, connection *entities.MCPConnection, reason string) error {
	securityRepo, ok := s.connections.(repositories.IMCPConnectionSecurityRepository)
	if ok {
		_, _ = securityRepo.UpdateOAuth(ctx, repositories.MCPConnectionOAuthUpdateInput{TenantID: connection.TenantID, ProjectID: connection.ProjectID, ID: connection.ID, CredentialReference: connection.OAuthAccessTokenReference, RefreshTokenReference: connection.OAuthRefreshTokenReference, CredentialStatus: entities.MCPCredentialActionRequired, OAuthStatus: entities.MCPOAuthActionRequired, OAuthExpiresAt: connection.OAuthExpiresAt, UpdatedBy: "system"})
	}
	return errors.New(reason)
}

func (s *MCPOAuthService) findOAuthConnection(ctx context.Context, tenantID, projectID, connectionID string) (*entities.MCPConnection, error) {
	if err := validateToolCatalogScope(tenantID, projectID, connectionID); err != nil {
		return nil, err
	}
	return s.connections.FindByID(ctx, tenantID, projectID, connectionID)
}

func (s *MCPOAuthService) auditOAuth(ctx context.Context, tenantID, actor string, connection *entities.MCPConnection, action entities.AuditAction) {
	if s.audit != nil && connection != nil {
		s.audit.Write(ctx, tenantID, actor, action, "mcp_connection", connection.ID, nil, map[string]any{"connectionId": connection.ID, "projectId": connection.ProjectID, "status": connection.OAuthStatus}, nil)
	}
}

func updatedEntity(connection *entities.MCPConnection) *entities.MCPConnection { return connection }

func randomOAuthValue(size int) (string, error) {
	bytes := make([]byte, size)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func hashOAuthValue(value string) []byte { return hashOAuthBytes([]byte(value)) }
func hashOAuthBytes(value []byte) []byte { sum := sha256.Sum256(value); return sum[:] }
