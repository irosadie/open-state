package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/irosadie/open-state/api/internal/application/dtos"
	domaincap "github.com/irosadie/open-state/api/internal/domain/capability"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	domain "github.com/irosadie/open-state/go-shared/domain"
)

// CapabilityService orchestrates the Capability Registry, its bindings, and
// sandbox test-invocation for the admin API (PRD §174). Every operation is
// tenant-scoped; the tenant is supplied by the caller from the authenticated
// context, never from the request body (PRD §4, §96).
//
// The test-invocation collaborators (provider resolver + schema validator) are
// injected so the application layer depends only on domain ports; the
// composition root wires the mock/sandbox implementation (PRD §2064).
type CapabilityService struct {
	repo             repositories.ICapabilityRepository
	providerResolver domaincap.ProviderResolver
	schemaValidator  domaincap.InputSchemaValidator
	audit            *AuditWriter
	rateLimiter      domaincap.RateLimiter
}

// NewCapabilityService builds a CapabilityService.
func NewCapabilityService(repo repositories.ICapabilityRepository, providerResolver domaincap.ProviderResolver, schemaValidator domaincap.InputSchemaValidator, audit *AuditWriter, rateLimiter domaincap.RateLimiter) *CapabilityService {
	return &CapabilityService{repo: repo, providerResolver: providerResolver, schemaValidator: schemaValidator, audit: audit, rateLimiter: rateLimiter}
}

// Create registers a new capability for the tenant (PRD §59).
func (s *CapabilityService) Create(ctx context.Context, tenantID string, req dtos.CreateCapabilityRequest) (*dtos.CapabilityDTO, error) {
	providerType, err := parseProviderType(req.ProviderType)
	if err != nil {
		return nil, err
	}
	if req.Name == "" {
		return nil, domain.NewValidation("name is required")
	}
	if err := validateProviderMapping(providerType, req.ProviderID, req.ProviderTool); err != nil {
		return nil, err
	}

	input, err := marshalSchema(req.InputSchema)
	if err != nil {
		return nil, domain.NewValidation("inputSchema must be valid JSON")
	}
	output, err := marshalSchema(req.OutputSchema)
	if err != nil {
		return nil, domain.NewValidation("outputSchema must be valid JSON")
	}

	version := req.Version
	if version == 0 {
		version = 1
	}
	credRef := optional(req.CredentialReference)

	cap, err := s.repo.Create(ctx, tenantID, req.Name, optional(req.Description), providerType, optional(req.ProviderID), optional(req.ProviderTool), input, output, version, credRef)
	if err != nil {
		return nil, err
	}
	return toCapabilityDTO(cap), nil
}

// FindByID returns a single tenant-scoped capability.
func (s *CapabilityService) FindByID(ctx context.Context, tenantID, id string) (*dtos.CapabilityDTO, error) {
	cap, err := s.repo.FindByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	return toCapabilityDTO(cap), nil
}

// List returns the tenant-scoped registry, optionally filtered by provider type
// and status (PRD §59).
func (s *CapabilityService) List(ctx context.Context, tenantID, providerType, status string) (*dtos.CapabilityListDTO, error) {
	list, err := s.repo.ListByTenantFiltered(ctx, tenantID, entities.ProviderType(providerType), entities.CapabilityStatus(status))
	if err != nil {
		return nil, err
	}
	data := make([]dtos.CapabilityDTO, 0, len(list))
	for i := range list {
		data = append(data, *toCapabilityDTO(&list[i]))
	}
	return &dtos.CapabilityListDTO{Data: data}, nil
}

// Update applies mutable field updates to a tenant-scoped capability. Missing
// fields are preserved (partial/PATCH semantics): only provided fields are
// overwritten.
func (s *CapabilityService) Update(ctx context.Context, tenantID, id string, req dtos.UpdateCapabilityRequest) (*dtos.CapabilityDTO, error) {
	existing, err := s.repo.FindByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}

	providerType := existing.ProviderType
	if req.ProviderType != "" {
		providerType, err = parseProviderType(req.ProviderType)
		if err != nil {
			return nil, err
		}
	}

	status := existing.Status
	if req.Status != "" {
		status, err = parseCapabilityStatus(req.Status)
		if err != nil {
			return nil, err
		}
	}

	input := existing.InputSchema
	if req.InputSchema != nil {
		input, err = marshalSchema(req.InputSchema)
		if err != nil {
			return nil, domain.NewValidation("inputSchema must be valid JSON")
		}
	}
	output := existing.OutputSchema
	if req.OutputSchema != nil {
		output, err = marshalSchema(req.OutputSchema)
		if err != nil {
			return nil, domain.NewValidation("outputSchema must be valid JSON")
		}
	}

	version := existing.Version
	if req.Version > 0 {
		version = req.Version
	}

	description := existing.Description
	if req.Description != "" {
		description = sql.NullString{String: req.Description, Valid: true}
	}
	providerID := existing.ProviderID
	if req.ProviderID != "" {
		providerID = sql.NullString{String: req.ProviderID, Valid: true}
	}
	providerTool := existing.ProviderTool
	if req.ProviderTool != "" {
		providerTool = sql.NullString{String: req.ProviderTool, Valid: true}
	}
	credRef := existing.CredentialReference
	if req.CredentialReference != "" {
		credRef = sql.NullString{String: req.CredentialReference, Valid: true}
	}
	if err := validateProviderMapping(providerType, providerID.String, providerTool.String); err != nil {
		return nil, err
	}

	cap, err := s.repo.Update(ctx, tenantID, id, nilStringPtr(description), providerType, nilStringPtr(providerID), nilStringPtr(providerTool), input, output, status, version, nilStringPtr(credRef))
	if err != nil {
		return nil, err
	}
	return toCapabilityDTO(cap), nil
}

// Delete disables a capability and its bindings for the tenant (PRD §59).
func (s *CapabilityService) Delete(ctx context.Context, tenantID, id string) error {
	_, err := s.repo.Disable(ctx, tenantID, id)
	return err
}

// ListBindings returns the bindings of a tenant-scoped capability.
func (s *CapabilityService) ListBindings(ctx context.Context, tenantID, capabilityID string) ([]dtos.CapabilityBindingDTO, error) {
	bindings, err := s.repo.ListBindingsByCapability(ctx, tenantID, capabilityID)
	if err != nil {
		return nil, err
	}
	out := make([]dtos.CapabilityBindingDTO, 0, len(bindings))
	for i := range bindings {
		out = append(out, *toBindingDTO(&bindings[i]))
	}
	return out, nil
}

// Bind scopes a capability to a tenant/workflow/state level (PRD §60). On
// success it appends a binding.created audit entry (PRD 50).
func (s *CapabilityService) Bind(ctx context.Context, tenantID, capabilityID, actor string, req dtos.CreateBindingRequest) (*dtos.CapabilityBindingDTO, error) {
	scopeType, err := parseScopeType(req.ScopeType)
	if err != nil {
		return nil, err
	}
	permission, err := parsePermission(req.Permission)
	if err != nil {
		return nil, err
	}
	if req.ScopeID == "" {
		return nil, domain.NewValidation("scopeId is required")
	}

	binding, err := s.repo.Bind(ctx, tenantID, capabilityID, scopeType, req.ScopeID, permission)
	if err != nil {
		return nil, err
	}
	if s.audit != nil {
		s.audit.Write(ctx, tenantID, actor, entities.AuditActionBindingCreated, "binding", binding.ID,
			nil, map[string]any{"capabilityId": capabilityID, "scopeType": scopeType, "scopeId": req.ScopeID, "permission": permission}, nil)
	}
	return toBindingDTO(binding), nil
}

// Unbind removes a tenant-scoped binding. On success it appends a
// binding.deleted audit entry (PRD 50).
func (s *CapabilityService) Unbind(ctx context.Context, tenantID, bindingID, actor string) error {
	if s.audit != nil {
		s.audit.Write(ctx, tenantID, actor, entities.AuditActionBindingDeleted, "binding", bindingID, nil, nil, nil)
	}
	return s.repo.Unbind(ctx, tenantID, bindingID)
}

// TestInvoke runs a capability through the security chain in mock/sandbox mode
// and returns the normalized result (PRD §2064, §153). It never invokes a real
// provider — the mock provider resolver is forced. On success it appends a
// capability.invoked audit entry; on denial/error a capability.denied entry
// (PRD 50).
func (s *CapabilityService) TestInvoke(ctx context.Context, tenantID, capabilityID, actor string, req dtos.TestInvocationRequest) (*dtos.TestInvocationResultDTO, error) {
	cap, err := s.repo.FindByID(ctx, tenantID, capabilityID)
	if err != nil {
		return nil, err
	}

	invoker := domaincap.NewCapabilityInvoker(
		domaincap.NewCapabilityResolver(s.repo),
		s.providerResolver,
		s.schemaValidator,
		s.rateLimiter,
		domaincap.NewInMemoryIdempotencyStore(),
	)

	inv := domaincap.Invocation{
		TenantID:     tenantID,
		Name:         cap.Name,
		Payload:      req.Payload,
		StateID:      req.ScopeID,
		CapabilityID: cap.ID,
		Policy:       domaincap.CapabilityPolicy{Timeout: 10 * time.Second},
	}

	result, err := invoker.Execute(ctx, inv)
	if err != nil {
		if s.audit != nil {
			s.audit.Write(ctx, tenantID, actor, entities.AuditActionCapabilityDenied, "capability", capabilityID,
				nil, map[string]any{"reason": err.Error()}, nil)
		}
		var ce *domaincap.CapabilityError
		if errors.As(err, &ce) {
			// Return the classified capability error as-is so the HTTP boundary
			// can surface kind/code to callers (PRD §87, §2951). Raw provider
			// errors are never exposed.
			return nil, ce
		}
		return nil, domain.NewInternal("test invocation failed")
	}

	if s.audit != nil {
		s.audit.Write(ctx, tenantID, actor, entities.AuditActionCapabilityInvoked, "capability", capabilityID, nil, nil, nil)
	}

	return &dtos.TestInvocationResultDTO{
		Data:       result.Data,
		FromMock:   result.FromMock,
		DurationMS: result.Duration.Milliseconds(),
		Event:      result.CapabilityEvent,
	}, nil
}

// ---- mapping & validation helpers ----

func toCapabilityDTO(c *entities.Capability) *dtos.CapabilityDTO {
	out := &dtos.CapabilityDTO{
		ID:                  c.ID,
		TenantID:            c.TenantID,
		Name:                c.Name,
		ProviderType:        string(c.ProviderType),
		Status:              string(c.Status),
		Version:             c.Version,
		CredentialReference: nilStringPtr(c.CredentialReference),
		InputSchema:         rawToAny(c.InputSchema),
		OutputSchema:        rawToAny(c.OutputSchema),
		CreatedAt:           c.CreatedAt.Format(time.RFC3339),
		UpdatedAt:           c.UpdatedAt.Format(time.RFC3339),
	}
	if c.Description.Valid {
		out.Description = &c.Description.String
	}
	if c.ProviderID.Valid {
		out.ProviderID = &c.ProviderID.String
	}
	if c.ProviderTool.Valid {
		out.ProviderTool = &c.ProviderTool.String
	}
	return out
}

func toBindingDTO(b *entities.CapabilityBinding) *dtos.CapabilityBindingDTO {
	return &dtos.CapabilityBindingDTO{
		ID:           b.ID,
		TenantID:     b.TenantID,
		CapabilityID: b.CapabilityID,
		ScopeType:    string(b.ScopeType),
		ScopeID:      b.ScopeID,
		Permission:   string(b.Permission),
		CreatedAt:    b.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    b.UpdatedAt.Format(time.RFC3339),
	}
}

func parseProviderType(s string) (entities.ProviderType, error) {
	if s == "" {
		return "", domain.NewValidation("providerType is required")
	}
	switch entities.ProviderType(s) {
	case entities.ProviderTypeMCP, entities.ProviderTypeInternal, entities.ProviderTypeHTTP, entities.ProviderTypeFuture:
		return entities.ProviderType(s), nil
	default:
		return "", domain.NewValidation("invalid providerType")
	}
}

func validateProviderMapping(providerType entities.ProviderType, providerServer, providerTool string) error {
	if providerType != entities.ProviderTypeMCP {
		return nil
	}
	if providerServer == "" || providerTool == "" {
		return domain.NewValidation("MCP capabilities require providerId (server alias) and providerTool")
	}
	if containsEndpoint(providerServer) || containsEndpoint(providerTool) {
		return domain.NewValidation("MCP provider mapping accepts a server alias and tool name, not an endpoint")
	}
	return nil
}

func containsEndpoint(value string) bool {
	return strings.Contains(value, "://")
}

func parseCapabilityStatus(s string) (entities.CapabilityStatus, error) {
	if s == "" {
		return entities.CapabilityActive, nil
	}
	switch entities.CapabilityStatus(s) {
	case entities.CapabilityActive, entities.CapabilityInactive, entities.CapabilityDisabled:
		return entities.CapabilityStatus(s), nil
	default:
		return "", domain.NewValidation("invalid status")
	}
}

func parseScopeType(s string) (entities.BindingScopeType, error) {
	switch entities.BindingScopeType(s) {
	case entities.BindingScopeTenant, entities.BindingScopeWorkflow, entities.BindingScopeState:
		return entities.BindingScopeType(s), nil
	default:
		return "", domain.NewValidation("invalid scopeType")
	}
}

func parsePermission(s string) (entities.BindingPermission, error) {
	if s == "" {
		return entities.BindingPermissionAllow, nil
	}
	switch entities.BindingPermission(s) {
	case entities.BindingPermissionAllow, entities.BindingPermissionDeny:
		return entities.BindingPermission(s), nil
	default:
		return "", domain.NewValidation("invalid permission")
	}
}

func marshalSchema(v any) ([]byte, error) {
	if v == nil {
		return []byte(`{}`), nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return []byte(`{}`), nil
	}
	return b, nil
}

func rawToAny(raw json.RawMessage) any {
	if len(raw) == 0 {
		return nil
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil
	}
	return v
}

func optional(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func nilStringPtr(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}
