package services

import (
	"context"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/irosadie/open-state/api/internal/application/dtos"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	domainsvc "github.com/irosadie/open-state/api/internal/domain/services"
	domain "github.com/irosadie/open-state/go-shared/domain"
)

var mcpAliasPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{1,127}$`)

type MCPConnectionService struct {
	repo     repositories.IMCPConnectionRepository
	projects repositories.IProjectRepository
	tester   domainsvc.MCPConnectionTester
	audit    *AuditWriter
	secrets  domainsvc.SecretStore
}

func NewMCPConnectionService(repo repositories.IMCPConnectionRepository, projects repositories.IProjectRepository, tester domainsvc.MCPConnectionTester, audit *AuditWriter, stores ...domainsvc.SecretStore) *MCPConnectionService {
	var secrets domainsvc.SecretStore
	if len(stores) > 0 {
		secrets = stores[0]
	}
	return &MCPConnectionService{repo: repo, projects: projects, tester: tester, audit: audit, secrets: secrets}
}

func (s *MCPConnectionService) List(ctx context.Context, tenantID, projectID string) (*dtos.MCPConnectionListDTO, error) {
	if err := validateScope(tenantID, projectID); err != nil {
		return nil, err
	}
	if err := s.ensureProject(ctx, tenantID, projectID); err != nil {
		return nil, err
	}
	items, err := s.repo.ListByProject(ctx, tenantID, projectID)
	if err != nil {
		return nil, err
	}
	out := make([]dtos.MCPConnectionDTO, 0, len(items))
	for i := range items {
		out = append(out, toMCPConnectionDTO(&items[i]))
	}
	return &dtos.MCPConnectionListDTO{Data: out}, nil
}

func (s *MCPConnectionService) Get(ctx context.Context, tenantID, projectID, id string) (*dtos.MCPConnectionDTO, error) {
	if err := validateScope(tenantID, projectID); err != nil {
		return nil, err
	}
	if _, err := uuid.Parse(id); err != nil {
		return nil, domain.NewValidation("invalid MCP connection id")
	}
	item, err := s.repo.FindByID(ctx, tenantID, projectID, id)
	if err != nil {
		return nil, err
	}
	dto := toMCPConnectionDTO(item)
	return &dto, nil
}

func (s *MCPConnectionService) Create(ctx context.Context, tenantID, projectID, actor string, req dtos.CreateMCPConnectionRequest) (*dtos.MCPConnectionDTO, error) {
	if err := validateScope(tenantID, projectID); err != nil {
		return nil, err
	}
	if err := s.ensureProject(ctx, tenantID, projectID); err != nil {
		return nil, err
	}
	config, err := normalizeMCPConnection(req)
	if err != nil {
		return nil, err
	}
	if req.CredentialValue != "" {
		if config.authType != entities.MCPAuthBearer {
			return nil, domain.NewValidation("credentialValue is only valid for bearer authentication")
		}
		if s.secrets == nil {
			return nil, domain.NewInternal("secret store is not configured")
		}
		ref, err := s.secrets.Put(ctx, tenantID, projectID, "mcp_bearer", req.CredentialValue)
		if err != nil {
			return nil, domain.NewInternal("MCP credential could not be stored")
		}
		config.credentialReference = &ref
		config.credentialStatus = entities.MCPCredentialConfigured
	}
	if req.OAuthClientSecretValue != "" {
		if config.authType != entities.MCPAuthOAuth {
			return nil, domain.NewValidation("oauthClientSecretValue requires OAuth authentication")
		}
		if s.secrets == nil {
			return nil, domain.NewInternal("secret store is not configured")
		}
		ref, err := s.secrets.Put(ctx, tenantID, projectID, "mcp_oauth_client_secret", req.OAuthClientSecretValue)
		if err != nil {
			return nil, domain.NewInternal("OAuth client secret could not be stored")
		}
		config.oauthClientSecretReference = &ref
	}
	item, err := s.repo.Create(ctx, repositories.MCPConnectionCreateInput{TenantID: tenantID, ProjectID: projectID, Name: config.name, Alias: config.alias, Transport: config.transport, Endpoint: config.endpoint, StdioProfile: config.stdioProfile, StdioArgs: config.stdioArgs, AuthType: config.authType, CredentialReference: config.credentialReference, CredentialStatus: config.credentialStatus, OAuthAuthorizationEndpoint: config.oauthAuthorizationEndpoint, OAuthTokenEndpoint: config.oauthTokenEndpoint, OAuthClientID: config.oauthClientID, OAuthClientSecretReference: config.oauthClientSecretReference, OAuthScopes: config.oauthScopes, OAuthRedirectURI: config.oauthRedirectURI, TimeoutMS: config.timeoutMS, MaxConcurrency: config.maxConcurrency, RateLimitPerSecond: config.rateLimitPerSecond, RateLimitBurst: config.rateLimitBurst, RetryMax: config.retryMax, CircuitFailureThreshold: config.circuitFailureThreshold, CircuitRecoverySeconds: config.circuitRecoverySeconds, Status: entities.MCPConnectionEnabled, CreatedBy: actor})
	if err != nil {
		return nil, err
	}
	s.auditMutation(ctx, tenantID, actor, entities.AuditActionMCPConnectionCreated, item, nil)
	dto := toMCPConnectionDTO(item)
	return &dto, nil
}

func (s *MCPConnectionService) Update(ctx context.Context, tenantID, projectID, id, actor string, req dtos.UpdateMCPConnectionRequest) (*dtos.MCPConnectionDTO, error) {
	if err := validateScope(tenantID, projectID); err != nil {
		return nil, err
	}
	if _, err := uuid.Parse(id); err != nil {
		return nil, domain.NewValidation("invalid MCP connection id")
	}
	existing, err := s.repo.FindByID(ctx, tenantID, projectID, id)
	if err != nil {
		return nil, err
	}
	config, err := normalizeMCPConnection(req)
	if err != nil {
		return nil, err
	}
	if existing.CredentialReference != nil && existing.AuthType == entities.MCPAuthBearer && config.authType != entities.MCPAuthBearer && s.secrets != nil {
		_ = s.secrets.Revoke(ctx, tenantID, projectID, *existing.CredentialReference)
	}
	if config.authType == entities.MCPAuthNone {
		config.credentialReference = nil
	} else if req.CredentialValue != "" {
		if config.authType != entities.MCPAuthBearer || s.secrets == nil {
			return nil, domain.NewValidation("credentialValue requires a configured bearer secret store")
		}
		if existing.CredentialReference != nil {
			ref, rotateErr := s.secrets.Rotate(ctx, tenantID, projectID, *existing.CredentialReference, "mcp_bearer", req.CredentialValue)
			if rotateErr != nil {
				return nil, domain.NewInternal("MCP credential could not be rotated")
			}
			config.credentialReference = &ref
		} else {
			ref, putErr := s.secrets.Put(ctx, tenantID, projectID, "mcp_bearer", req.CredentialValue)
			if putErr != nil {
				return nil, domain.NewInternal("MCP credential could not be stored")
			}
			config.credentialReference = &ref
		}
	} else if config.credentialReference == nil && ((config.authType == entities.MCPAuthBearer && existing.AuthType == entities.MCPAuthBearer) || (config.authType == entities.MCPAuthOAuth && existing.AuthType == entities.MCPAuthOAuth)) {
		config.credentialReference = existing.CredentialReference
	}
	if config.authType != entities.MCPAuthOAuth {
		if existing.OAuthClientSecretReference != nil && s.secrets != nil {
			_ = s.secrets.Revoke(ctx, tenantID, projectID, *existing.OAuthClientSecretReference)
		}
		config.oauthClientSecretReference = nil
	} else if req.OAuthClientSecretValue != "" {
		if s.secrets == nil {
			return nil, domain.NewInternal("secret store is not configured")
		}
		if existing.OAuthClientSecretReference != nil {
			ref, rotateErr := s.secrets.Rotate(ctx, tenantID, projectID, *existing.OAuthClientSecretReference, "mcp_oauth_client_secret", req.OAuthClientSecretValue)
			if rotateErr != nil {
				return nil, domain.NewInternal("OAuth client secret could not be rotated")
			}
			config.oauthClientSecretReference = &ref
		} else {
			ref, putErr := s.secrets.Put(ctx, tenantID, projectID, "mcp_oauth_client_secret", req.OAuthClientSecretValue)
			if putErr != nil {
				return nil, domain.NewInternal("OAuth client secret could not be stored")
			}
			config.oauthClientSecretReference = &ref
		}
	} else {
		config.oauthClientSecretReference = existing.OAuthClientSecretReference
	}
	config.credentialStatus = credentialStatus(config.authType, config.credentialReference)
	item, err := s.repo.Update(ctx, repositories.MCPConnectionUpdateInput{TenantID: tenantID, ProjectID: projectID, ID: id, Name: config.name, Alias: config.alias, Transport: config.transport, Endpoint: config.endpoint, StdioProfile: config.stdioProfile, StdioArgs: config.stdioArgs, AuthType: config.authType, CredentialReference: config.credentialReference, CredentialStatus: config.credentialStatus, OAuthAuthorizationEndpoint: config.oauthAuthorizationEndpoint, OAuthTokenEndpoint: config.oauthTokenEndpoint, OAuthClientID: config.oauthClientID, OAuthClientSecretReference: config.oauthClientSecretReference, OAuthScopes: config.oauthScopes, OAuthRedirectURI: config.oauthRedirectURI, TimeoutMS: config.timeoutMS, MaxConcurrency: config.maxConcurrency, RateLimitPerSecond: config.rateLimitPerSecond, RateLimitBurst: config.rateLimitBurst, RetryMax: config.retryMax, CircuitFailureThreshold: config.circuitFailureThreshold, CircuitRecoverySeconds: config.circuitRecoverySeconds, UpdatedBy: actor})
	if err != nil {
		return nil, err
	}
	s.auditMutation(ctx, tenantID, actor, entities.AuditActionMCPConnectionUpdated, item, safeMCPConnection(existing))
	dto := toMCPConnectionDTO(item)
	return &dto, nil
}

func (s *MCPConnectionService) Delete(ctx context.Context, tenantID, projectID, id, actor string) error {
	if err := validateScope(tenantID, projectID); err != nil {
		return err
	}
	if err := validateConnectionID(id); err != nil {
		return err
	}
	item, err := s.repo.FindByID(ctx, tenantID, projectID, id)
	if err != nil {
		return err
	}
	if err := s.repo.Delete(ctx, tenantID, projectID, id); err != nil {
		return err
	}
	if s.secrets != nil {
		for _, ref := range []*string{item.CredentialReference, item.OAuthAccessTokenReference, item.OAuthRefreshTokenReference, item.OAuthClientSecretReference} {
			if ref != nil {
				_ = s.secrets.Revoke(ctx, tenantID, projectID, *ref)
			}
		}
	}
	s.auditMutation(ctx, tenantID, actor, entities.AuditActionMCPConnectionDeleted, item, nil)
	return nil
}

func (s *MCPConnectionService) SetStatus(ctx context.Context, tenantID, projectID, id, actor string, status entities.MCPConnectionStatus) (*dtos.MCPConnectionDTO, error) {
	if err := validateScope(tenantID, projectID); err != nil {
		return nil, err
	}
	if err := validateConnectionID(id); err != nil {
		return nil, err
	}
	if status != entities.MCPConnectionEnabled && status != entities.MCPConnectionDisabled {
		return nil, domain.NewValidation("invalid MCP connection status")
	}
	item, err := s.repo.UpdateStatus(ctx, tenantID, projectID, id, status, actor)
	if err != nil {
		return nil, err
	}
	action := entities.AuditActionMCPConnectionEnabled
	if status == entities.MCPConnectionDisabled {
		action = entities.AuditActionMCPConnectionDisabled
	}
	s.auditMutation(ctx, tenantID, actor, action, item, nil)
	dto := toMCPConnectionDTO(item)
	return &dto, nil
}

func (s *MCPConnectionService) Test(ctx context.Context, tenantID, projectID, id, actor string) (*dtos.MCPConnectionDTO, error) {
	if err := validateScope(tenantID, projectID); err != nil {
		return nil, err
	}
	if err := validateConnectionID(id); err != nil {
		return nil, err
	}
	item, err := s.repo.FindByID(ctx, tenantID, projectID, id)
	if err != nil {
		return nil, err
	}
	if item.Status == entities.MCPConnectionDisabled {
		return nil, domain.NewConflict("disabled MCP connection cannot be tested")
	}
	if s.tester == nil {
		return nil, domain.NewInternal("MCP connection tester is not configured")
	}
	result, testErr := s.tester.Handshake(ctx, item)
	status := entities.MCPTestReady
	errorCode := result.ErrorCode
	if testErr != nil || !result.Ready {
		status = entities.MCPTestFailed
		if errorCode == "" {
			errorCode = "mcp_handshake_failed"
		}
	}
	updated, err := s.repo.RecordTest(ctx, tenantID, projectID, id, status, errorCode, actor)
	if err != nil {
		return nil, err
	}
	if healthRepo, ok := s.repo.(repositories.IMCPConnectionSecurityRepository); ok {
		healthStatus := entities.MCPHealthHealthy
		healthReason := ""
		if testErr != nil || !result.Ready {
			healthStatus = entities.MCPHealthUnavailable
			healthReason = errorCode
		}
		_, _ = healthRepo.RecordHealth(ctx, repositories.MCPConnectionHealthUpdateInput{
			TenantID: tenantID, ProjectID: projectID, ID: id, HealthStatus: healthStatus,
			HealthReason: optionalString(healthReason), LastSuccessAt: successTimeForHealth(healthStatus),
			ConsecutiveFailures: healthFailuresForStatus(healthStatus), Actor: actor,
		})
		if refreshed, refreshErr := s.repo.FindByID(ctx, tenantID, projectID, id); refreshErr == nil {
			updated = refreshed
		}
	}
	s.auditMutation(ctx, tenantID, actor, entities.AuditActionMCPConnectionTested, updated, map[string]any{"result": string(status), "errorCode": errorCode})
	dto := toMCPConnectionDTO(updated)
	return &dto, nil
}

func (s *MCPConnectionService) RotateCredential(ctx context.Context, tenantID, projectID, id, actor, value string) (*dtos.MCPConnectionDTO, error) {
	if err := validateScope(tenantID, projectID); err != nil {
		return nil, err
	}
	if err := validateConnectionID(id); err != nil {
		return nil, err
	}
	if strings.TrimSpace(value) == "" {
		return nil, domain.NewValidation("credential value is required")
	}
	connection, err := s.repo.FindByID(ctx, tenantID, projectID, id)
	if err != nil {
		return nil, err
	}
	if connection.AuthType != entities.MCPAuthBearer {
		return nil, domain.NewValidation("credential rotation is only available for bearer authentication")
	}
	securityRepo, ok := s.repo.(repositories.IMCPConnectionSecurityRepository)
	if !ok || s.secrets == nil {
		return nil, domain.NewInternal("MCP credential lifecycle is not configured")
	}
	var ref string
	if connection.CredentialReference != nil && strings.TrimSpace(*connection.CredentialReference) != "" {
		ref, err = s.secrets.Rotate(ctx, tenantID, projectID, *connection.CredentialReference, "mcp_bearer", value)
	} else {
		ref, err = s.secrets.Put(ctx, tenantID, projectID, "mcp_bearer", value)
	}
	if err != nil {
		return nil, domain.NewInternal("MCP credential could not be rotated")
	}
	updated, err := securityRepo.UpdateCredential(ctx, repositories.MCPConnectionCredentialUpdateInput{
		TenantID: tenantID, ProjectID: projectID, ID: id, CredentialReference: &ref,
		CredentialStatus: entities.MCPCredentialConfigured, UpdatedBy: actor,
	})
	if err != nil {
		return nil, err
	}
	s.auditMutation(ctx, tenantID, actor, entities.AuditActionMCPCredentialRotated, updated, map[string]any{"credentialStatus": entities.MCPCredentialConfigured})
	dto := toMCPConnectionDTO(updated)
	return &dto, nil
}

func (s *MCPConnectionService) RevokeCredential(ctx context.Context, tenantID, projectID, id, actor string) (*dtos.MCPConnectionDTO, error) {
	if err := validateScope(tenantID, projectID); err != nil {
		return nil, err
	}
	if err := validateConnectionID(id); err != nil {
		return nil, err
	}
	connection, err := s.repo.FindByID(ctx, tenantID, projectID, id)
	if err != nil {
		return nil, err
	}
	if connection.AuthType != entities.MCPAuthBearer {
		return nil, domain.NewValidation("credential revocation is only available for bearer authentication")
	}
	securityRepo, ok := s.repo.(repositories.IMCPConnectionSecurityRepository)
	if !ok || s.secrets == nil {
		return nil, domain.NewInternal("MCP credential lifecycle is not configured")
	}
	if connection.CredentialReference != nil {
		if err := s.secrets.Revoke(ctx, tenantID, projectID, *connection.CredentialReference); err != nil {
			return nil, domain.NewConflict("MCP credential cannot be revoked by the configured secret store")
		}
	}
	updated, err := securityRepo.UpdateCredential(ctx, repositories.MCPConnectionCredentialUpdateInput{
		TenantID: tenantID, ProjectID: projectID, ID: id, CredentialStatus: entities.MCPCredentialMissing, UpdatedBy: actor,
	})
	if err != nil {
		return nil, err
	}
	s.auditMutation(ctx, tenantID, actor, entities.AuditActionMCPCredentialRevoked, updated, map[string]any{"credentialStatus": entities.MCPCredentialMissing})
	dto := toMCPConnectionDTO(updated)
	return &dto, nil
}

func (s *MCPConnectionService) CredentialStatus(ctx context.Context, tenantID, projectID, id string) (*dtos.MCPCredentialStatusDTO, error) {
	if err := validateScope(tenantID, projectID); err != nil {
		return nil, err
	}
	if err := validateConnectionID(id); err != nil {
		return nil, err
	}
	connection, err := s.repo.FindByID(ctx, tenantID, projectID, id)
	if err != nil {
		return nil, err
	}
	status := string(connection.CredentialStatus)
	if connection.AuthType == entities.MCPAuthNone {
		status = string(domainsvc.SecretStatusConfigured)
	}
	if connection.CredentialReference != nil && s.secrets != nil {
		secretStatus, statusErr := s.secrets.Status(ctx, tenantID, projectID, *connection.CredentialReference)
		if statusErr != nil {
			return nil, domain.NewInternal("MCP credential status is unavailable")
		}
		status = string(secretStatus)
	}
	return &dtos.MCPCredentialStatusDTO{Status: status, CredentialStatus: string(connection.CredentialStatus), CanRotate: connection.AuthType == entities.MCPAuthBearer, CanRevoke: connection.AuthType == entities.MCPAuthBearer && connection.CredentialReference != nil}, nil
}

func (s *MCPConnectionService) Diagnose(ctx context.Context, tenantID, projectID, id, actor string) (*dtos.MCPConnectionDTO, error) {
	result, err := s.Test(ctx, tenantID, projectID, id, actor)
	if err != nil {
		return nil, err
	}
	if s.audit != nil {
		s.audit.Write(ctx, tenantID, actor, entities.AuditActionMCPHealthDiagnosed, "mcp_connection", id, nil, map[string]any{"projectId": projectID, "healthStatus": result.HealthStatus, "lastTestStatus": result.LastTestStatus}, nil)
	}
	return result, nil
}

func (s *MCPConnectionService) ResetHealth(ctx context.Context, tenantID, projectID, id, actor string) (*dtos.MCPConnectionDTO, error) {
	if err := validateScope(tenantID, projectID); err != nil {
		return nil, err
	}
	if err := validateConnectionID(id); err != nil {
		return nil, err
	}
	securityRepo, ok := s.repo.(repositories.IMCPConnectionSecurityRepository)
	if !ok {
		return nil, domain.NewInternal("MCP health lifecycle is not configured")
	}
	updated, err := securityRepo.ResetHealth(ctx, tenantID, projectID, id, actor)
	if err != nil {
		return nil, err
	}
	s.auditMutation(ctx, tenantID, actor, entities.AuditActionMCPHealthReset, updated, map[string]any{"healthStatus": entities.MCPHealthUnknown})
	dto := toMCPConnectionDTO(updated)
	return &dto, nil
}

type normalizedMCPConnection struct {
	name, alias                                                                                          string
	transport                                                                                            entities.MCPConnectionTransport
	endpoint, stdioProfile, credentialReference                                                          *string
	stdioArgs                                                                                            []string
	authType                                                                                             entities.MCPConnectionAuthType
	credentialStatus                                                                                     entities.MCPConnectionCredentialStatus
	oauthAuthorizationEndpoint, oauthTokenEndpoint, oauthClientID, oauthRedirectURI                      *string
	oauthClientSecretReference                                                                           *string
	oauthScopes                                                                                          []string
	timeoutMS, maxConcurrency, rateLimitBurst, retryMax, circuitFailureThreshold, circuitRecoverySeconds int
	rateLimitPerSecond                                                                                   float64
}

func normalizeMCPConnection(req dtos.CreateMCPConnectionRequest) (normalizedMCPConnection, error) {
	name, alias := strings.TrimSpace(req.Name), strings.ToLower(strings.TrimSpace(req.Alias))
	if name == "" {
		return normalizedMCPConnection{}, domain.NewValidation("name is required")
	}
	if alias == "" || !mcpAliasPattern.MatchString(alias) {
		return normalizedMCPConnection{}, domain.NewValidation("alias must use 2-128 lowercase letters, numbers, dots, hyphens, or underscores")
	}
	transport := entities.MCPConnectionTransport(strings.ToLower(strings.TrimSpace(req.Transport)))
	if transport != entities.MCPTransportStreamableHTTP && transport != entities.MCPTransportSSE && transport != entities.MCPTransportSTDIO {
		return normalizedMCPConnection{}, domain.NewValidation("transport must be streamable_http, sse, or stdio")
	}
	authType := entities.MCPConnectionAuthType(strings.ToLower(strings.TrimSpace(req.AuthType)))
	if authType != entities.MCPAuthNone && authType != entities.MCPAuthBearer && authType != entities.MCPAuthOAuth {
		return normalizedMCPConnection{}, domain.NewValidation("authType must be none, bearer, or oauth")
	}

	endpointValue := strings.TrimSpace(req.Endpoint)
	profileValue := strings.TrimSpace(req.StdioProfile)
	if transport == entities.MCPTransportSTDIO {
		if profileValue == "" {
			return normalizedMCPConnection{}, domain.NewValidation("stdioProfile is required for stdio transport")
		}
		if endpointValue != "" {
			return normalizedMCPConnection{}, domain.NewValidation("endpoint is not valid for stdio transport")
		}
	} else {
		if endpointValue == "" {
			return normalizedMCPConnection{}, domain.NewValidation("endpoint is required for remote MCP transport")
		}
		u, err := url.ParseRequestURI(endpointValue)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			return normalizedMCPConnection{}, domain.NewValidation("endpoint must be a valid http or https URL")
		}
		profileValue = ""
	}
	if authType == entities.MCPAuthNone && strings.TrimSpace(req.CredentialReference) != "" {
		return normalizedMCPConnection{}, domain.NewValidation("credentialReference is only valid for bearer or oauth authentication")
	}
	if req.CredentialValue != "" && strings.TrimSpace(req.CredentialReference) != "" {
		return normalizedMCPConnection{}, domain.NewValidation("send either credentialReference or credentialValue, not both")
	}
	if authType != entities.MCPAuthOAuth && (strings.TrimSpace(req.OAuthAuthorizationEndpoint) != "" || strings.TrimSpace(req.OAuthTokenEndpoint) != "" || strings.TrimSpace(req.OAuthClientID) != "" || strings.TrimSpace(req.OAuthRedirectURI) != "" || len(req.OAuthScopes) > 0) {
		return normalizedMCPConnection{}, domain.NewValidation("OAuth settings require oauth authentication")
	}
	if authType == entities.MCPAuthOAuth {
		if strings.TrimSpace(req.OAuthAuthorizationEndpoint) == "" || strings.TrimSpace(req.OAuthTokenEndpoint) == "" || strings.TrimSpace(req.OAuthClientID) == "" || strings.TrimSpace(req.OAuthRedirectURI) == "" {
			return normalizedMCPConnection{}, domain.NewValidation("OAuth authorization endpoint, token endpoint, client id, and redirect URI are required")
		}
		for _, endpoint := range []string{req.OAuthAuthorizationEndpoint, req.OAuthTokenEndpoint} {
			u, err := url.ParseRequestURI(strings.TrimSpace(endpoint))
			if err != nil || u.Scheme != "https" || u.Host == "" {
				return normalizedMCPConnection{}, domain.NewValidation("OAuth endpoints must use HTTPS")
			}
		}
		if len(req.OAuthScopes) > 32 {
			return normalizedMCPConnection{}, domain.NewValidation("oauthScopes cannot contain more than 32 values")
		}
	}
	credentialRef := strings.TrimSpace(req.CredentialReference)
	var credentialReference *string
	if credentialRef != "" {
		credentialReference = &credentialRef
	}
	timeoutMS := req.TimeoutMS
	if timeoutMS == 0 {
		timeoutMS = 10000
	}
	maxConcurrency := req.MaxConcurrency
	if maxConcurrency == 0 {
		maxConcurrency = 4
	}
	rateLimitPerSecond := req.RateLimitPerSecond
	if rateLimitPerSecond == 0 {
		rateLimitPerSecond = 10
	}
	rateLimitBurst := req.RateLimitBurst
	if rateLimitBurst == 0 {
		rateLimitBurst = 20
	}
	retryMax := req.RetryMax
	if retryMax == 0 {
		retryMax = 1
	}
	circuitFailureThreshold := req.CircuitFailureThreshold
	if circuitFailureThreshold == 0 {
		circuitFailureThreshold = 5
	}
	circuitRecoverySeconds := req.CircuitRecoverySeconds
	if circuitRecoverySeconds == 0 {
		circuitRecoverySeconds = 30
	}
	if timeoutMS < 100 || timeoutMS > 300000 || maxConcurrency < 1 || maxConcurrency > 256 || rateLimitPerSecond <= 0 || rateLimitPerSecond > 10000 || rateLimitBurst < 1 || rateLimitBurst > 10000 || retryMax < 0 || retryMax > 5 || circuitFailureThreshold < 1 || circuitFailureThreshold > 100 || circuitRecoverySeconds < 1 || circuitRecoverySeconds > 86400 {
		return normalizedMCPConnection{}, domain.NewValidation("MCP resilience policy is outside the allowed range")
	}
	return normalizedMCPConnection{name: name, alias: alias, transport: transport, endpoint: optionalString(endpointValue), stdioProfile: optionalString(profileValue), stdioArgs: append([]string{}, req.StdioArgs...), authType: authType, credentialReference: credentialReference, credentialStatus: credentialStatus(authType, credentialReference), oauthAuthorizationEndpoint: optionalString(strings.TrimSpace(req.OAuthAuthorizationEndpoint)), oauthTokenEndpoint: optionalString(strings.TrimSpace(req.OAuthTokenEndpoint)), oauthClientID: optionalString(strings.TrimSpace(req.OAuthClientID)), oauthScopes: append([]string{}, req.OAuthScopes...), oauthRedirectURI: optionalString(strings.TrimSpace(req.OAuthRedirectURI)), timeoutMS: timeoutMS, maxConcurrency: maxConcurrency, rateLimitPerSecond: rateLimitPerSecond, rateLimitBurst: rateLimitBurst, retryMax: retryMax, circuitFailureThreshold: circuitFailureThreshold, circuitRecoverySeconds: circuitRecoverySeconds}, nil
}

func credentialStatus(auth entities.MCPConnectionAuthType, ref *string) entities.MCPConnectionCredentialStatus {
	if auth == entities.MCPAuthNone {
		return entities.MCPCredentialConfigured
	}
	if ref == nil || strings.TrimSpace(*ref) == "" {
		return entities.MCPCredentialMissing
	}
	return entities.MCPCredentialConfigured
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func (s *MCPConnectionService) ensureProject(ctx context.Context, tenantID, projectID string) error {
	if s.projects == nil {
		return domain.NewInternal("project repository is not configured")
	}
	_, err := s.projects.FindByID(ctx, tenantID, projectID)
	return err
}

func validateScope(tenantID, projectID string) error {
	if _, err := uuid.Parse(tenantID); err != nil {
		return domain.NewValidation("invalid tenant id")
	}
	if _, err := uuid.Parse(projectID); err != nil {
		return domain.NewValidation("invalid project id")
	}
	return nil
}

func validateConnectionID(id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return domain.NewValidation("invalid MCP connection id")
	}
	return nil
}

func toMCPConnectionDTO(item *entities.MCPConnection) dtos.MCPConnectionDTO {
	var lastTestedAt *string
	if item.LastTestedAt != nil {
		value := item.LastTestedAt.UTC().Format(time.RFC3339)
		lastTestedAt = &value
	}
	return dtos.MCPConnectionDTO{ID: item.ID, TenantID: item.TenantID, ProjectID: item.ProjectID, Name: item.Name, Alias: item.Alias, Transport: string(item.Transport), Endpoint: item.Endpoint, StdioProfile: item.StdioProfile, StdioArgs: append([]string{}, item.StdioArgs...), AuthType: string(item.AuthType), CredentialStatus: string(item.CredentialStatus), OAuthAuthorizationEndpoint: item.OAuthAuthorizationEndpoint, OAuthTokenEndpoint: item.OAuthTokenEndpoint, OAuthClientID: item.OAuthClientID, OAuthScopes: append([]string{}, item.OAuthScopes...), OAuthRedirectURI: item.OAuthRedirectURI, OAuthStatus: string(item.OAuthStatus), Status: string(item.Status), LastTestStatus: string(item.LastTestStatus), LastTestErrorCode: item.LastTestErrorCode, LastTestedAt: lastTestedAt, HealthStatus: string(item.HealthStatus), HealthReason: item.HealthReason, LastSuccessAt: formatMCPTimePtr(item.LastSuccessAt), ConsecutiveFailures: item.ConsecutiveFailures, CircuitOpenedAt: formatMCPTimePtr(item.CircuitOpenedAt), TimeoutMS: item.TimeoutMS, MaxConcurrency: item.MaxConcurrency, RateLimitPerSecond: item.RateLimitPerSecond, RateLimitBurst: item.RateLimitBurst, RetryMax: item.RetryMax, CircuitFailureThreshold: item.CircuitFailureThreshold, CircuitRecoverySeconds: item.CircuitRecoverySeconds, CreatedAt: item.CreatedAt.UTC().Format(time.RFC3339), UpdatedAt: item.UpdatedAt.UTC().Format(time.RFC3339)}
}

func formatMCPTimePtr(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := value.UTC().Format(time.RFC3339)
	return &formatted
}

func successTimeForHealth(status entities.MCPConnectionHealthStatus) *time.Time {
	if status != entities.MCPHealthHealthy {
		return nil
	}
	now := time.Now().UTC()
	return &now
}

func healthFailuresForStatus(status entities.MCPConnectionHealthStatus) int {
	if status == entities.MCPHealthHealthy {
		return 0
	}
	return 1
}

func (s *MCPConnectionService) auditMutation(ctx context.Context, tenantID, actor string, action entities.AuditAction, item *entities.MCPConnection, before any) {
	if s.audit == nil {
		return
	}
	s.audit.Write(ctx, tenantID, actor, action, "mcp_connection", item.ID, before, safeMCPConnection(item), nil)
}

func safeMCPConnection(item *entities.MCPConnection) map[string]any {
	return map[string]any{"id": item.ID, "projectId": item.ProjectID, "alias": item.Alias, "transport": item.Transport, "authType": item.AuthType, "credentialStatus": item.CredentialStatus, "oauthStatus": item.OAuthStatus, "healthStatus": item.HealthStatus, "status": item.Status, "lastTestStatus": item.LastTestStatus}
}
