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
}

func NewMCPConnectionService(repo repositories.IMCPConnectionRepository, projects repositories.IProjectRepository, tester domainsvc.MCPConnectionTester, audit *AuditWriter) *MCPConnectionService {
	return &MCPConnectionService{repo: repo, projects: projects, tester: tester, audit: audit}
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
	item, err := s.repo.Create(ctx, repositories.MCPConnectionCreateInput{TenantID: tenantID, ProjectID: projectID, Name: config.name, Alias: config.alias, Transport: config.transport, Endpoint: config.endpoint, StdioProfile: config.stdioProfile, StdioArgs: config.stdioArgs, AuthType: config.authType, CredentialReference: config.credentialReference, CredentialStatus: config.credentialStatus, Status: entities.MCPConnectionEnabled, CreatedBy: actor})
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
	if config.authType == entities.MCPAuthNone {
		config.credentialReference = nil
	} else if config.credentialReference == nil {
		config.credentialReference = existing.CredentialReference
	}
	config.credentialStatus = credentialStatus(config.authType, config.credentialReference)
	item, err := s.repo.Update(ctx, repositories.MCPConnectionUpdateInput{TenantID: tenantID, ProjectID: projectID, ID: id, Name: config.name, Alias: config.alias, Transport: config.transport, Endpoint: config.endpoint, StdioProfile: config.stdioProfile, StdioArgs: config.stdioArgs, AuthType: config.authType, CredentialReference: config.credentialReference, CredentialStatus: config.credentialStatus, UpdatedBy: actor})
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
	s.auditMutation(ctx, tenantID, actor, entities.AuditActionMCPConnectionTested, updated, map[string]any{"result": string(status), "errorCode": errorCode})
	dto := toMCPConnectionDTO(updated)
	return &dto, nil
}

type normalizedMCPConnection struct {
	name, alias                                 string
	transport                                   entities.MCPConnectionTransport
	endpoint, stdioProfile, credentialReference *string
	stdioArgs                                   []string
	authType                                    entities.MCPConnectionAuthType
	credentialStatus                            entities.MCPConnectionCredentialStatus
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
	credentialRef := strings.TrimSpace(req.CredentialReference)
	var credentialReference *string
	if credentialRef != "" {
		credentialReference = &credentialRef
	}
	return normalizedMCPConnection{name: name, alias: alias, transport: transport, endpoint: optionalString(endpointValue), stdioProfile: optionalString(profileValue), stdioArgs: append([]string{}, req.StdioArgs...), authType: authType, credentialReference: credentialReference, credentialStatus: credentialStatus(authType, credentialReference)}, nil
}

func credentialStatus(auth entities.MCPConnectionAuthType, ref *string) entities.MCPConnectionCredentialStatus {
	if auth == entities.MCPAuthNone {
		return entities.MCPCredentialConfigured
	}
	if ref == nil || strings.TrimSpace(*ref) == "" {
		return entities.MCPCredentialMissing
	}
	if auth == entities.MCPAuthOAuth {
		return entities.MCPCredentialActionRequired
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
	return dtos.MCPConnectionDTO{ID: item.ID, TenantID: item.TenantID, ProjectID: item.ProjectID, Name: item.Name, Alias: item.Alias, Transport: string(item.Transport), Endpoint: item.Endpoint, StdioProfile: item.StdioProfile, StdioArgs: append([]string{}, item.StdioArgs...), AuthType: string(item.AuthType), CredentialStatus: string(item.CredentialStatus), Status: string(item.Status), LastTestStatus: string(item.LastTestStatus), LastTestErrorCode: item.LastTestErrorCode, LastTestedAt: lastTestedAt, CreatedAt: item.CreatedAt.UTC().Format(time.RFC3339), UpdatedAt: item.UpdatedAt.UTC().Format(time.RFC3339)}
}

func (s *MCPConnectionService) auditMutation(ctx context.Context, tenantID, actor string, action entities.AuditAction, item *entities.MCPConnection, before any) {
	if s.audit == nil {
		return
	}
	s.audit.Write(ctx, tenantID, actor, action, "mcp_connection", item.ID, before, safeMCPConnection(item), nil)
}

func safeMCPConnection(item *entities.MCPConnection) map[string]any {
	return map[string]any{"id": item.ID, "projectId": item.ProjectID, "alias": item.Alias, "transport": item.Transport, "authType": item.AuthType, "credentialStatus": item.CredentialStatus, "status": item.Status, "lastTestStatus": item.LastTestStatus}
}
