package services

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/irosadie/open-state/api/internal/application/dtos"
	"github.com/irosadie/open-state/api/internal/domain/engine"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	domain "github.com/irosadie/open-state/go-shared/domain"
)

// ProjectCapabilityMCPBindingService owns the explicit mapping between a
// logical capability used in a workflow and a discovered provider tool in the
// current project. It is also the authoritative validator for publication and
// the safe resolver consumed by State MCP.
type ProjectCapabilityMCPBindingService struct {
	bindings     repositories.IProjectCapabilityMCPBindingRepository
	capabilities repositories.ICapabilityRepository
	connections  repositories.IMCPConnectionRepository
	catalog      repositories.IMCPToolCatalogRepository
	projects     repositories.IProjectRepository
	audit        *AuditWriter
}

func NewProjectCapabilityMCPBindingService(
	bindings repositories.IProjectCapabilityMCPBindingRepository,
	capabilities repositories.ICapabilityRepository,
	connections repositories.IMCPConnectionRepository,
	catalog repositories.IMCPToolCatalogRepository,
	projects repositories.IProjectRepository,
	audit *AuditWriter,
) *ProjectCapabilityMCPBindingService {
	return &ProjectCapabilityMCPBindingService{bindings: bindings, capabilities: capabilities, connections: connections, catalog: catalog, projects: projects, audit: audit}
}

func (s *ProjectCapabilityMCPBindingService) ListOptions(ctx context.Context, tenantID, projectID string) (*dtos.MCPToolOptionListDTO, error) {
	if err := validateProjectMCPBindingScope(tenantID, projectID); err != nil {
		return nil, err
	}
	if err := s.ensureProject(ctx, tenantID, projectID); err != nil {
		return nil, err
	}
	items, err := s.bindings.ListEligibleToolOptions(ctx, tenantID, projectID)
	if err != nil {
		return nil, err
	}
	data := make([]dtos.MCPToolOptionDTO, 0, len(items))
	for i := range items {
		data = append(data, toMCPToolOptionDTO(&items[i]))
	}
	return &dtos.MCPToolOptionListDTO{Data: data}, nil
}

func (s *ProjectCapabilityMCPBindingService) List(ctx context.Context, tenantID, projectID string) (*dtos.ProjectCapabilityMCPBindingListDTO, error) {
	if err := validateProjectMCPBindingScope(tenantID, projectID); err != nil {
		return nil, err
	}
	if err := s.ensureProject(ctx, tenantID, projectID); err != nil {
		return nil, err
	}
	items, err := s.bindings.ListByProject(ctx, tenantID, projectID)
	if err != nil {
		return nil, err
	}
	data := make([]dtos.ProjectCapabilityMCPBindingDTO, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for i := range items {
		data = append(data, toProjectCapabilityMCPBindingDTO(&items[i]))
		seen[items[i].CapabilityID] = struct{}{}
	}

	// Missing mappings are returned explicitly so the builder and State MCP do
	// not silently fall back to legacy provider metadata.
	capabilities, err := s.capabilities.ListByTenantFiltered(ctx, tenantID, entities.ProviderTypeMCP, entities.CapabilityActive)
	if err != nil {
		return nil, err
	}
	for i := range capabilities {
		capability := &capabilities[i]
		if _, exists := seen[capability.ID]; exists {
			continue
		}
		data = append(data, missingMCPBindingDTO(tenantID, projectID, capability))
	}
	sort.Slice(data, func(i, j int) bool { return data[i].CapabilityName < data[j].CapabilityName })
	return &dtos.ProjectCapabilityMCPBindingListDTO{Data: data}, nil
}

func (s *ProjectCapabilityMCPBindingService) Upsert(ctx context.Context, tenantID, projectID, capabilityID, actor string, req dtos.UpsertProjectCapabilityMCPBindingRequest) (*dtos.ProjectCapabilityMCPBindingDTO, error) {
	if err := validateProjectMCPBindingScope(tenantID, projectID); err != nil {
		return nil, err
	}
	if _, err := uuid.Parse(capabilityID); err != nil {
		return nil, domain.NewValidation("invalid capability id")
	}
	if _, err := uuid.Parse(req.ConnectionID); err != nil {
		return nil, domain.NewValidation("connectionId must be a valid UUID")
	}
	if _, err := uuid.Parse(req.ToolID); err != nil {
		return nil, domain.NewValidation("toolId must be a valid UUID")
	}
	if err := s.ensureProject(ctx, tenantID, projectID); err != nil {
		return nil, err
	}
	capability, err := s.capabilities.FindByID(ctx, tenantID, capabilityID)
	if err != nil {
		return nil, err
	}
	if capability.ID != capabilityID || capability.TenantID != tenantID {
		return nil, domain.NewNotFound("capability not found")
	}
	if capability.ProviderType != entities.ProviderTypeMCP {
		return nil, domain.NewValidation("only MCP capabilities can have a project MCP binding")
	}
	if capability.Status != entities.CapabilityActive {
		return nil, domain.NewConflict("inactive MCP capabilities cannot be bound")
	}
	connection, err := s.connections.FindByID(ctx, tenantID, projectID, req.ConnectionID)
	if err != nil {
		return nil, err
	}
	if connection.ID != req.ConnectionID || connection.TenantID != tenantID || connection.ProjectID != projectID {
		return nil, domain.NewNotFound("MCP connection not found")
	}
	if connection.Status != entities.MCPConnectionEnabled {
		return nil, domain.NewConflict("disabled MCP connections cannot be bound")
	}
	catalog, err := s.catalog.Get(ctx, tenantID, projectID, req.ConnectionID)
	if err != nil {
		return nil, err
	}
	var selected *entities.MCPDiscoveredTool
	for i := range catalog.Tools {
		if catalog.Tools[i].ID == req.ToolID {
			selected = &catalog.Tools[i]
			break
		}
	}
	if selected == nil {
		return nil, domain.NewNotFound("MCP tool is not in this connection's catalog")
	}
	if selected.TenantID != "" && selected.TenantID != tenantID || selected.ProjectID != "" && selected.ProjectID != projectID || selected.ConnectionID != "" && selected.ConnectionID != req.ConnectionID {
		return nil, domain.NewNotFound("MCP tool does not belong to this connection")
	}
	if selected.Availability != entities.MCPToolPresent || !selected.Enabled {
		return nil, domain.NewConflict("only present and enabled MCP tools can be bound")
	}
	if err := s.bindings.Upsert(ctx, repositories.ProjectCapabilityMCPBindingUpsertInput{
		TenantID: tenantID, ProjectID: projectID, CapabilityID: capabilityID,
		ConnectionID: req.ConnectionID, ToolID: req.ToolID, ToolFingerprint: selected.Fingerprint,
	}); err != nil {
		return nil, err
	}
	binding, err := s.bindings.FindByCapability(ctx, tenantID, projectID, capabilityID)
	if err != nil {
		return nil, err
	}
	if s.audit != nil {
		s.audit.Write(ctx, tenantID, actor, entities.AuditActionBindingCreated, "project_capability_mcp_binding", binding.ID, nil, map[string]string{
			"projectId": projectID, "capabilityId": capabilityID, "connectionAlias": connection.Alias,
			"toolName": selected.Name, "toolFingerprint": selected.Fingerprint,
		}, nil)
	}
	dto := toProjectCapabilityMCPBindingDTO(binding)
	return &dto, nil
}

func (s *ProjectCapabilityMCPBindingService) Delete(ctx context.Context, tenantID, projectID, capabilityID, actor string) error {
	if err := validateProjectMCPBindingScope(tenantID, projectID); err != nil {
		return err
	}
	if _, err := uuid.Parse(capabilityID); err != nil {
		return domain.NewValidation("invalid capability id")
	}
	if err := s.bindings.Delete(ctx, tenantID, projectID, capabilityID); err != nil {
		return err
	}
	if s.audit != nil {
		s.audit.Write(ctx, tenantID, actor, entities.AuditActionBindingDeleted, "project_capability_mcp_binding", capabilityID, nil, map[string]string{
			"projectId": projectID, "capabilityId": capabilityID,
		}, nil)
	}
	return nil
}

// Resolve returns the project binding used by State MCP. Callers must treat a
// missing binding as a hard mapping failure, not as permission to use legacy
// provider fields.
func (s *ProjectCapabilityMCPBindingService) Resolve(ctx context.Context, tenantID, projectID, capabilityID string) (*entities.ProjectCapabilityMCPBinding, error) {
	if err := validateProjectMCPBindingScope(tenantID, projectID); err != nil {
		return nil, err
	}
	return s.bindings.FindByCapability(ctx, tenantID, projectID, capabilityID)
}

// ValidateWorkflow is the authoritative publication check for project MCP
// bindings. It returns safe, actionable issue details without provider secrets.
func (s *ProjectCapabilityMCPBindingService) ValidateWorkflow(ctx context.Context, tenantID, projectID string, raw []byte) (*domain.DomainError, error) {
	var definition engine.WorkflowDefinition
	if err := json.Unmarshal(raw, &definition); err != nil {
		return domain.NewValidation("workflow definition is invalid"), nil
	}
	issues := make([]dtos.WorkflowValidationIssue, 0)
	seen := make(map[string]struct{})
	for _, node := range definition.Nodes {
		for _, capabilityName := range node.Capabilities {
			name := strings.TrimSpace(capabilityName)
			if name == "" {
				continue
			}
			capability, err := s.capabilities.FindByName(ctx, tenantID, name)
			if err != nil {
				var notFound *domain.DomainError
				if errors.As(err, &notFound) && notFound.Code == domain.ErrNotFound {
					issues = appendMCPBindingIssue(issues, seen, "MCP_CAPABILITY_NOT_FOUND", "MCP capability "+name+" is not registered", node.ID)
					continue
				}
				return nil, err
			}
			if capability.ProviderType != entities.ProviderTypeMCP {
				continue
			}
			binding, err := s.bindings.FindByCapability(ctx, tenantID, projectID, capability.ID)
			if err != nil {
				var notFound *domain.DomainError
				if errors.As(err, &notFound) && notFound.Code == domain.ErrNotFound {
					issues = appendMCPBindingIssue(issues, seen, "MCP_BINDING_MISSING", "MCP capability "+name+" has no project tool binding", node.ID)
					continue
				}
				return nil, err
			}
			if binding.Health != entities.ProjectCapabilityMCPBindingActive {
				message := "MCP capability " + name + " is unavailable"
				if binding.HealthReason != "" {
					message += ": " + binding.HealthReason
				}
				issues = appendMCPBindingIssue(issues, seen, "MCP_BINDING_UNAVAILABLE", message, node.ID)
			}
		}
	}
	if len(issues) == 0 {
		return nil, nil
	}
	details, err := json.Marshal(issues)
	if err != nil {
		return domain.NewValidation("workflow MCP bindings are invalid"), nil
	}
	return domain.NewValidationWithDetails("workflow has unavailable MCP bindings", details), nil
}

func appendMCPBindingIssue(issues []dtos.WorkflowValidationIssue, seen map[string]struct{}, code, message, nodeID string) []dtos.WorkflowValidationIssue {
	key := code + "|" + nodeID + "|" + message
	if _, exists := seen[key]; exists {
		return issues
	}
	seen[key] = struct{}{}
	return append(issues, dtos.WorkflowValidationIssue{Code: code, Message: message, NodeID: nodeID})
}

func (s *ProjectCapabilityMCPBindingService) ensureProject(ctx context.Context, tenantID, projectID string) error {
	if s.projects == nil {
		return domain.NewInternal("project repository is not configured")
	}
	_, err := s.projects.FindByID(ctx, tenantID, projectID)
	return err
}

func validateProjectMCPBindingScope(tenantID, projectID string) error {
	if _, err := uuid.Parse(tenantID); err != nil {
		return domain.NewValidation("invalid tenant id")
	}
	if _, err := uuid.Parse(projectID); err != nil {
		return domain.NewValidation("invalid project id")
	}
	return nil
}

func toMCPToolOptionDTO(item *entities.ProjectMCPToolOption) dtos.MCPToolOptionDTO {
	return dtos.MCPToolOptionDTO{ConnectionID: item.ConnectionID, ConnectionName: item.ConnectionName, ConnectionAlias: item.ConnectionAlias, ConnectionStatus: string(item.ConnectionStatus), ToolID: item.ToolID, ToolName: item.ToolName, ToolTitle: item.ToolTitle, ToolDescription: item.ToolDescription, InputSchema: append(json.RawMessage(nil), item.InputSchema...), ToolFingerprint: item.ToolFingerprint}
}

func toProjectCapabilityMCPBindingDTO(item *entities.ProjectCapabilityMCPBinding) dtos.ProjectCapabilityMCPBindingDTO {
	toolEnabled := item.ToolEnabled
	return dtos.ProjectCapabilityMCPBindingDTO{ID: item.ID, TenantID: item.TenantID, ProjectID: item.ProjectID, CapabilityID: item.CapabilityID, CapabilityName: item.CapabilityName, CapabilityDescription: item.CapabilityDescription, ConnectionID: item.ConnectionID, ConnectionName: item.ConnectionName, ConnectionAlias: item.ConnectionAlias, ConnectionStatus: string(item.ConnectionStatus), ToolID: item.ToolID, ToolName: item.ToolName, ToolTitle: item.ToolTitle, ToolDescription: item.ToolDescription, BoundToolFingerprint: item.BoundToolFingerprint, CurrentToolFingerprint: item.CurrentToolFingerprint, ToolEnabled: &toolEnabled, ToolAvailability: string(item.ToolAvailability), ToolDriftStatus: string(item.ToolDriftStatus), Health: string(item.Health), HealthReason: item.HealthReason, CreatedAt: item.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"), UpdatedAt: item.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00")}
}

func missingMCPBindingDTO(tenantID, projectID string, capability *entities.Capability) dtos.ProjectCapabilityMCPBindingDTO {
	var description *string
	if capability.Description.Valid {
		description = &capability.Description.String
	}
	return dtos.ProjectCapabilityMCPBindingDTO{TenantID: tenantID, ProjectID: projectID, CapabilityID: capability.ID, CapabilityName: capability.Name, CapabilityDescription: description, Health: string(entities.ProjectCapabilityMCPBindingMissingMapping), HealthReason: "No project MCP tool binding configured"}
}
