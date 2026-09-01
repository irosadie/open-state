package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/irosadie/open-state/api/internal/application/dtos"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	domainsvc "github.com/irosadie/open-state/api/internal/domain/services"
	domain "github.com/irosadie/open-state/go-shared/domain"
)

const (
	maxDiscoveredTools      = 512
	maxToolDescriptionBytes = 8 * 1024
	maxToolSchemaBytes      = 64 * 1024
	maxToolAnnotationsBytes = 16 * 1024
	maxCatalogPayloadBytes  = 4 * 1024 * 1024
)

var sensitiveMetadataPattern = regexp.MustCompile(`(?i)(bearer\s+[a-z0-9._~+/=-]{12,}|sk-[a-z0-9_-]{12,}|(?:api[_ -]?key|secret|token|password)\s*[:=]\s*[a-z0-9._~+/=-]{8,})`)

// MCPToolCatalogService owns explicit discovery and catalog reconciliation. It
// never invokes a provider business tool; the discoverer port exposes only
// initialize plus tools/list.
type MCPToolCatalogService struct {
	connections repositories.IMCPConnectionRepository
	catalog     repositories.IMCPToolCatalogRepository
	discoverer  domainsvc.MCPToolDiscoverer
	audit       *AuditWriter
}

func NewMCPToolCatalogService(connections repositories.IMCPConnectionRepository, catalog repositories.IMCPToolCatalogRepository, discoverer domainsvc.MCPToolDiscoverer, audit *AuditWriter) *MCPToolCatalogService {
	return &MCPToolCatalogService{connections: connections, catalog: catalog, discoverer: discoverer, audit: audit}
}

func (s *MCPToolCatalogService) List(ctx context.Context, tenantID, projectID, connectionID string) (*dtos.MCPToolCatalogDTO, error) {
	if err := validateToolCatalogScope(tenantID, projectID, connectionID); err != nil {
		return nil, err
	}
	if _, err := s.connections.FindByID(ctx, tenantID, projectID, connectionID); err != nil {
		return nil, err
	}
	result, err := s.catalog.Get(ctx, tenantID, projectID, connectionID)
	if err != nil {
		return nil, err
	}
	return toMCPToolCatalogDTO(result), nil
}

func (s *MCPToolCatalogService) Refresh(ctx context.Context, tenantID, projectID, connectionID, actor string) (*dtos.MCPToolCatalogDTO, error) {
	if err := validateToolCatalogScope(tenantID, projectID, connectionID); err != nil {
		return nil, err
	}
	connection, err := s.connections.FindByID(ctx, tenantID, projectID, connectionID)
	if err != nil {
		return nil, err
	}
	if connection.Status == entities.MCPConnectionDisabled {
		return nil, domain.NewConflict("disabled MCP connection cannot be discovered")
	}
	if s.discoverer == nil {
		return nil, domain.NewInternal("MCP tool discoverer is not configured")
	}
	result, discoverErr := s.discoverer.DiscoverTools(ctx, connection)
	if discoverErr != nil || result.ErrorCode != "" {
		code := result.ErrorCode
		if code == "" {
			code = "mcp_discovery_failed"
		}
		return nil, s.recordDiscoveryFailure(ctx, tenantID, projectID, connectionID, actor, code)
	}
	tools, catalogFingerprint, err := sanitizeDiscoveredTools(result.Tools)
	if err != nil {
		code := "mcp_discovery_invalid_response"
		if errorsCode, ok := err.(discoveryValidationError); ok {
			code = errorsCode.Code
		}
		return nil, s.recordDiscoveryFailure(ctx, tenantID, projectID, connectionID, actor, code)
	}
	if _, err := s.catalog.Reconcile(ctx, repositories.MCPToolCatalogReconcileInput{
		TenantID: tenantID, ProjectID: projectID, ConnectionID: connectionID, Actor: actor,
		CatalogFingerprint: catalogFingerprint, Tools: tools,
	}); err != nil {
		return nil, err
	}
	s.auditCatalog(ctx, tenantID, actor, entities.AuditActionMCPToolCatalogRefreshed, connectionID, map[string]any{
		"projectId": projectID, "toolCount": len(tools), "catalogFingerprint": catalogFingerprint,
	})
	return s.List(ctx, tenantID, projectID, connectionID)
}

func (s *MCPToolCatalogService) SetEnabled(ctx context.Context, tenantID, projectID, connectionID, toolName, actor string, enabled bool) (*dtos.MCPDiscoveredToolDTO, error) {
	if err := validateToolCatalogScope(tenantID, projectID, connectionID); err != nil {
		return nil, err
	}
	if _, err := s.connections.FindByID(ctx, tenantID, projectID, connectionID); err != nil {
		return nil, err
	}
	toolName = strings.TrimSpace(toolName)
	if toolName == "" || len(toolName) > 255 {
		return nil, domain.NewValidation("toolName must be 1-255 characters")
	}
	tool, err := s.catalog.SetEnabled(ctx, tenantID, projectID, connectionID, toolName, enabled)
	if err != nil {
		return nil, err
	}
	action := entities.AuditActionMCPToolEnabled
	if !enabled {
		action = entities.AuditActionMCPToolDisabled
	}
	s.auditCatalog(ctx, tenantID, actor, action, tool.ID, map[string]any{
		"projectId": projectID, "connectionId": connectionID, "toolName": tool.Name,
		"enabled": enabled, "availability": tool.Availability, "fingerprint": tool.Fingerprint,
	})
	dto := toMCPDiscoveredToolDTO(tool)
	return &dto, nil
}

func (s *MCPToolCatalogService) recordDiscoveryFailure(ctx context.Context, tenantID, projectID, connectionID, actor, code string) error {
	if _, err := s.catalog.RecordFailure(ctx, repositories.MCPToolCatalogFailureInput{
		TenantID: tenantID, ProjectID: projectID, ConnectionID: connectionID, Actor: actor, ErrorCode: code,
	}); err != nil {
		return err
	}
	s.auditCatalog(ctx, tenantID, actor, entities.AuditActionMCPToolDiscoveryFailed, connectionID, map[string]any{
		"projectId": projectID, "connectionId": connectionID, "errorCode": code,
	})
	details, _ := json.Marshal(map[string]string{"code": code})
	return &domain.DomainError{Code: domain.ErrInternal, Message: "MCP tool discovery failed", Details: details}
}

func validateToolCatalogScope(tenantID, projectID, connectionID string) error {
	if _, err := uuid.Parse(tenantID); err != nil {
		return domain.NewValidation("invalid tenant id")
	}
	if _, err := uuid.Parse(projectID); err != nil {
		return domain.NewValidation("invalid project id")
	}
	if _, err := uuid.Parse(connectionID); err != nil {
		return domain.NewValidation("invalid MCP connection id")
	}
	return nil
}

type discoveryValidationError struct{ Code string }

func (e discoveryValidationError) Error() string { return e.Code }

func sanitizeDiscoveredTools(input []domainsvc.MCPToolDefinition) ([]repositories.MCPDiscoveredToolInput, string, error) {
	if len(input) > maxDiscoveredTools {
		return nil, "", discoveryValidationError{Code: "mcp_discovery_tool_limit_exceeded"}
	}
	seen := make(map[string]struct{}, len(input))
	tools := make([]repositories.MCPDiscoveredToolInput, 0, len(input))
	totalBytes := 0
	for _, item := range input {
		name := strings.TrimSpace(item.Name)
		if name == "" || len(name) > 255 {
			return nil, "", discoveryValidationError{Code: "mcp_discovery_invalid_tool_name"}
		}
		if _, exists := seen[name]; exists {
			return nil, "", discoveryValidationError{Code: "mcp_discovery_duplicate_tool"}
		}
		seen[name] = struct{}{}
		title := redactMetadata(strings.TrimSpace(item.Title))
		if len(title) > 255 {
			return nil, "", discoveryValidationError{Code: "mcp_discovery_title_limit_exceeded"}
		}
		description := redactMetadata(item.Description)
		if len(description) > maxToolDescriptionBytes {
			return nil, "", discoveryValidationError{Code: "mcp_discovery_description_limit_exceeded"}
		}
		inputSchema, err := sanitizeJSONMetadata(item.InputSchema, maxToolSchemaBytes)
		if err != nil {
			return nil, "", discoveryValidationError{Code: "mcp_discovery_schema_limit_exceeded"}
		}
		annotations, err := sanitizeJSONMetadata(item.Annotations, maxToolAnnotationsBytes)
		if err != nil {
			return nil, "", discoveryValidationError{Code: "mcp_discovery_annotations_limit_exceeded"}
		}
		fingerprint := fingerprintTool(name, title, description, inputSchema, annotations)
		totalBytes += len(name) + len(title) + len(description) + len(inputSchema) + len(annotations)
		if totalBytes > maxCatalogPayloadBytes {
			return nil, "", discoveryValidationError{Code: "mcp_discovery_payload_limit_exceeded"}
		}
		tools = append(tools, repositories.MCPDiscoveredToolInput{
			Name: name, Title: title, Description: description, InputSchema: inputSchema,
			Annotations: annotations, Fingerprint: fingerprint,
		})
	}
	ordered := append([]repositories.MCPDiscoveredToolInput(nil), tools...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Name < ordered[j].Name })
	digest := sha256.New()
	for _, tool := range ordered {
		digest.Write([]byte(tool.Name))
		digest.Write([]byte{0})
		digest.Write([]byte(tool.Fingerprint))
		digest.Write([]byte{0})
	}
	return tools, hex.EncodeToString(digest.Sum(nil)), nil
}

func sanitizeJSONMetadata(raw json.RawMessage, limit int) (json.RawMessage, error) {
	if len(raw) == 0 || len(raw) > limit || !json.Valid(raw) {
		return nil, discoveryValidationError{Code: "mcp_discovery_invalid_json"}
	}
	var value any // JSON Schema and MCP annotations are intentionally dynamic metadata.
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, err
	}
	if _, ok := value.(map[string]any); !ok {
		return nil, discoveryValidationError{Code: "mcp_discovery_metadata_must_be_object"}
	}
	redactJSONValue(value)
	result, err := json.Marshal(value)
	if err != nil || len(result) > limit {
		return nil, discoveryValidationError{Code: "mcp_discovery_metadata_limit_exceeded"}
	}
	return result, nil
}

func redactJSONValue(value any) {
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			if sensitiveMetadataKey(key) {
				typed[key] = redactSensitiveJSONSubtree(child)
				continue
			}
			if stringValue, ok := child.(string); ok {
				typed[key] = sensitiveMetadataPattern.ReplaceAllString(stringValue, "[REDACTED]")
				continue
			}
			redactJSONValue(child)
		}
	case []any:
		for _, child := range typed {
			redactJSONValue(child)
		}
	}
}

func redactSensitiveJSONSubtree(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			typed[key] = redactSensitiveJSONSubtree(child)
		}
		return typed
	case []any:
		for index, child := range typed {
			typed[index] = redactSensitiveJSONSubtree(child)
		}
		return typed
	case string:
		return "[REDACTED]"
	default:
		return value
	}
}

func redactMetadata(value string) string {
	return sensitiveMetadataPattern.ReplaceAllString(value, "[REDACTED]")
}

func sensitiveMetadataKey(key string) bool {
	key = strings.ToLower(key)
	return strings.Contains(key, "token") || strings.Contains(key, "secret") ||
		strings.Contains(key, "password") || strings.Contains(key, "api_key") ||
		strings.Contains(key, "apikey")
}

func fingerprintTool(name, title, description string, inputSchema, annotations json.RawMessage) string {
	digest := sha256.New()
	digest.Write([]byte(name))
	digest.Write([]byte{0})
	digest.Write([]byte(title))
	digest.Write([]byte{0})
	digest.Write([]byte(description))
	digest.Write([]byte{0})
	digest.Write(inputSchema)
	digest.Write([]byte{0})
	digest.Write(annotations)
	return hex.EncodeToString(digest.Sum(nil))
}

func toMCPToolCatalogDTO(catalog *entities.MCPToolCatalog) *dtos.MCPToolCatalogDTO {
	if catalog == nil {
		return &dtos.MCPToolCatalogDTO{Tools: []dtos.MCPDiscoveredToolDTO{}}
	}
	result := &dtos.MCPToolCatalogDTO{ConnectionID: catalog.ConnectionID, Tools: make([]dtos.MCPDiscoveredToolDTO, 0, len(catalog.Tools))}
	for i := range catalog.Tools {
		result.Tools = append(result.Tools, toMCPDiscoveredToolDTO(&catalog.Tools[i]))
	}
	if catalog.LatestRun != nil {
		run := toMCPDiscoveryRunDTO(catalog.LatestRun)
		result.LatestRun = &run
	}
	if catalog.LastSuccessfulRun != nil {
		run := toMCPDiscoveryRunDTO(catalog.LastSuccessfulRun)
		result.LastSuccessfulRun = &run
	}
	return result
}

func toMCPDiscoveryRunDTO(run *entities.MCPDiscoveryRun) dtos.MCPDiscoveryRunDTO {
	return dtos.MCPDiscoveryRunDTO{
		ID: run.ID, TenantID: run.TenantID, ProjectID: run.ProjectID, ConnectionID: run.ConnectionID,
		Status: string(run.Status), ToolCount: run.ToolCount, CatalogFingerprint: run.CatalogFingerprint,
		ErrorCode: run.ErrorCode, StartedAt: run.StartedAt.UTC().Format(time.RFC3339), CompletedAt: run.CompletedAt.UTC().Format(time.RFC3339), CreatedBy: run.CreatedBy,
	}
}

func toMCPDiscoveredToolDTO(tool *entities.MCPDiscoveredTool) dtos.MCPDiscoveredToolDTO {
	var removedAt *string
	if tool.RemovedAt != nil {
		value := tool.RemovedAt.UTC().Format(time.RFC3339)
		removedAt = &value
	}
	return dtos.MCPDiscoveredToolDTO{
		ID: tool.ID, TenantID: tool.TenantID, ProjectID: tool.ProjectID, ConnectionID: tool.ConnectionID,
		Name: tool.Name, Title: tool.Title, Description: tool.Description,
		InputSchema: append(json.RawMessage(nil), tool.InputSchema...), Annotations: append(json.RawMessage(nil), tool.Annotations...),
		Fingerprint: tool.Fingerprint, Enabled: tool.Enabled, Availability: string(tool.Availability), DriftStatus: string(tool.DriftStatus),
		FirstSeenAt: tool.FirstSeenAt.UTC().Format(time.RFC3339), LastSeenAt: tool.LastSeenAt.UTC().Format(time.RFC3339), RemovedAt: removedAt,
		DiscoveryRunID: tool.DiscoveryRunID, CreatedAt: tool.CreatedAt.UTC().Format(time.RFC3339), UpdatedAt: tool.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func (s *MCPToolCatalogService) auditCatalog(ctx context.Context, tenantID, actor string, action entities.AuditAction, resourceID string, after map[string]any) {
	if s.audit != nil {
		s.audit.Write(ctx, tenantID, actor, action, "mcp_tool_catalog", resourceID, nil, after, nil)
	}
}
