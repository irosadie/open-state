package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"

	domaincap "github.com/irosadie/open-state/api/internal/domain/capability"
	"github.com/irosadie/open-state/api/internal/domain/engine"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	domainservices "github.com/irosadie/open-state/api/internal/domain/services"
	domain "github.com/irosadie/open-state/go-shared/domain"
)

// MCPGatewayMode controls whether provider execution is enforced by OpenState
// or remains an advisory two-MCP compatibility flow.
type MCPGatewayMode string

const (
	MCPGatewayModeAdvisory MCPGatewayMode = "advisory"
	MCPGatewayModeSecure   MCPGatewayMode = "secure"
)

// GatewayRuntimeReader is the narrow runtime seam needed by the gateway. It
// derives the project, workflow, and current state from the persisted instance;
// none of those values are trusted from a provider request.
type GatewayRuntimeReader interface {
	GetCurrentState(ctx context.Context, tenantID, instanceID string) (*entities.WorkflowInstance, *entities.StateInstance, error)
	CurrentStateInfo(ctx context.Context, tenantID, instanceID string) (*engine.StateInfo, error)
}

// MCPGatewayProvider executes a target selected by the server-side binding
// resolver. The provider adapter owns transport, authentication, and MCP
// session details.
type MCPGatewayProvider interface {
	InvokeTool(ctx context.Context, connection *entities.MCPConnection, tool *entities.MCPDiscoveredTool, payload map[string]any, timeout time.Duration) (domainservices.MCPToolCallResult, error)
}

// GatewayInvocationRequest is the only input accepted by the secure gateway.
// It intentionally has no endpoint, provider alias, provider tool, or
// credential fields.
type GatewayInvocationRequest struct {
	TenantID       string
	InstanceID     string
	CapabilityName string
	Payload        map[string]any
	CorrelationID  string
	IdempotencyKey string
}

// GatewayInvocationResult is a safe State MCP projection. Provider connection
// details are retained only in evidence and never returned to the LLM.
type GatewayInvocationResult struct {
	InstanceID     string
	StateID        string
	CapabilityName string
	Data           map[string]any
	Status         entities.CapabilityEvidenceStatus
	Reused         bool
}

// MCPGatewayService performs the complete state-gated provider execution
// chain. A provider call is impossible until the current state, logical
// capability, project binding, connection, and discovered tool all validate.
type MCPGatewayService struct {
	runtime      GatewayRuntimeReader
	capabilities repositories.ICapabilityRepository
	bindings     repositories.IProjectCapabilityMCPBindingRepository
	connections  repositories.IMCPConnectionRepository
	catalog      repositories.IMCPToolCatalogRepository
	evidence     repositories.ICapabilityEvidenceRepository
	context      repositories.IContextRepository
	workflows    repositories.IWorkflowRepository
	provider     MCPGatewayProvider
	validator    domaincap.InputSchemaValidator
	fallback     *domaincap.CapabilityInvoker
	timeout      time.Duration
	locksMu      sync.Mutex
	locks        map[string]*sync.Mutex
}

func NewMCPGatewayService(
	runtime GatewayRuntimeReader,
	capabilities repositories.ICapabilityRepository,
	bindings repositories.IProjectCapabilityMCPBindingRepository,
	connections repositories.IMCPConnectionRepository,
	catalog repositories.IMCPToolCatalogRepository,
	evidence repositories.ICapabilityEvidenceRepository,
	contextRepo repositories.IContextRepository,
	workflows repositories.IWorkflowRepository,
	provider MCPGatewayProvider,
	validator domaincap.InputSchemaValidator,
	fallback *domaincap.CapabilityInvoker,
	timeout time.Duration,
) *MCPGatewayService {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &MCPGatewayService{
		runtime: runtime, capabilities: capabilities, bindings: bindings,
		connections: connections, catalog: catalog, evidence: evidence,
		context: contextRepo, workflows: workflows, provider: provider,
		validator: validator, fallback: fallback, timeout: timeout, locks: map[string]*sync.Mutex{},
	}
}

// Execute resolves all authorization and provider routing data internally,
// then invokes exactly the bound discovered tool. It fails closed on every
// missing or unhealthy link in the chain.
func (s *MCPGatewayService) Execute(ctx context.Context, req GatewayInvocationRequest) (*GatewayInvocationResult, error) {
	if s.runtime == nil || s.capabilities == nil {
		return nil, gatewayUnavailable("gateway dependencies are not configured")
	}
	if strings.TrimSpace(req.TenantID) == "" || strings.TrimSpace(req.InstanceID) == "" || strings.TrimSpace(req.CapabilityName) == "" {
		return nil, domaincap.NewCapabilityError(domaincap.ErrorKindValidation, "capability.gateway_input_invalid", "instance and capability are required")
	}
	if strings.TrimSpace(req.IdempotencyKey) == "" || strings.TrimSpace(req.CorrelationID) == "" {
		return nil, domaincap.NewCapabilityError(domaincap.ErrorKindValidation, "capability.gateway_input_invalid", "correlationId and idempotencyKey are required")
	}

	inst, stateInst, err := s.runtime.GetCurrentState(ctx, req.TenantID, req.InstanceID)
	if err != nil || inst == nil {
		return nil, gatewayUnauthorized("workflow instance is unavailable")
	}
	info, err := s.runtime.CurrentStateInfo(ctx, req.TenantID, req.InstanceID)
	if err != nil || info == nil {
		return nil, gatewayUnauthorized("current workflow state is unavailable")
	}
	stateID := strings.TrimSpace(info.StateID)
	if stateID == "" && stateInst != nil {
		stateID = stateInst.StateKey
		if stateID == "" && stateInst.StateID != nil {
			stateID = *stateInst.StateID
		}
	}
	if stateID == "" {
		return nil, gatewayUnauthorized("current workflow state is unavailable")
	}
	projectID := strings.TrimSpace(info.ProjectID)
	if projectID == "" && s.workflows != nil {
		if version, versionErr := s.workflows.FindCurrentVersionByWorkflow(ctx, req.TenantID, inst.WorkflowID); versionErr == nil {
			projectID = version.ProjectID
		}
	}
	if projectID == "" {
		return nil, gatewayUnauthorized("workflow project is unavailable")
	}
	if !declaredCapability(info.Capabilities, req.CapabilityName) {
		return nil, gatewayUnauthorized("capability is not declared by the current state")
	}

	resolver := domaincap.NewCapabilityResolver(s.capabilities)
	resolved, err := resolver.Resolve(ctx, req.TenantID, req.CapabilityName, inst.WorkflowID, stateID)
	if err != nil {
		return nil, err
	}
	if resolved.ProviderType != entities.ProviderTypeMCP {
		return s.executeNonMCP(ctx, req, inst, stateID, resolved)
	}
	if s.bindings == nil || s.connections == nil || s.catalog == nil || s.provider == nil || s.evidence == nil {
		return nil, gatewayUnavailable("MCP gateway dependencies are not configured")
	}
	if s.validator != nil && len(resolved.InputSchema) > 0 {
		if err := s.validator.Validate(req.Payload, resolved.InputSchema); err != nil {
			return nil, domaincap.NewCapabilityError(domaincap.ErrorKindValidation, "capability.validation_failed", "capability input is invalid")
		}
	}

	binding, err := s.bindings.FindByCapability(ctx, req.TenantID, projectID, resolved.ID)
	if err != nil {
		if isNotFound(err) {
			return nil, gatewayFailure("capability.mapping_missing", "capability has no project MCP binding")
		}
		return nil, gatewayFailure("capability.mapping_unavailable", "project MCP binding is unavailable")
	}
	if binding == nil || binding.Health != entities.ProjectCapabilityMCPBindingActive {
		return nil, gatewayFailure("capability.mapping_unavailable", "capability provider binding is unavailable")
	}
	connection, err := s.connections.FindByID(ctx, req.TenantID, projectID, binding.ConnectionID)
	if err != nil || connection == nil || connection.Status != entities.MCPConnectionEnabled {
		return nil, gatewayFailure("capability.provider_unavailable", "capability provider connection is unavailable")
	}
	catalogSnapshot, err := s.catalog.Get(ctx, req.TenantID, projectID, binding.ConnectionID)
	if err != nil || catalogSnapshot == nil {
		return nil, gatewayFailure("capability.catalog_unavailable", "capability provider catalog is unavailable")
	}
	tool := findBoundTool(catalogSnapshot.Tools, binding)
	if tool == nil || !tool.Enabled || tool.Availability != entities.MCPToolPresent || tool.Fingerprint != binding.BoundToolFingerprint {
		return nil, gatewayFailure("capability.tool_unavailable", "capability provider tool is unavailable")
	}

	release := s.lockIdempotency(req.TenantID, projectID, inst.ID, stateID, resolved.ID, req.IdempotencyKey)
	defer release()
	previous, findErr := s.evidence.FindByIdempotency(ctx, req.TenantID, projectID, inst.ID, stateID, resolved.ID, req.IdempotencyKey)
	if findErr == nil && previous != nil {
		if previous.Status != entities.CapabilityEvidenceSucceeded {
			return nil, domaincap.NewCapabilityError(domaincap.ErrorKindBusiness, "capability.idempotency_conflict", "idempotency key already has a failed outcome")
		}
		return &GatewayInvocationResult{
			InstanceID: inst.ID, StateID: stateID, CapabilityName: resolved.Name,
			Data: decodeEvidenceResult(previous.Result), Status: previous.Status, Reused: true,
		}, nil
	}
	if findErr != nil && !isNotFound(findErr) {
		return nil, gatewayUnavailable("capability idempotency state is unavailable")
	}

	callCtx := domainservices.WithMCPCallOptions(ctx, domainservices.MCPCallOptions{CorrelationID: req.CorrelationID, IdempotencyKey: req.IdempotencyKey, Idempotent: req.IdempotencyKey != ""})
	call, err := s.provider.InvokeTool(callCtx, connection, tool, req.Payload, s.timeout)
	if err != nil {
		s.recordFailedEvidence(ctx, req, projectID, stateID, resolved, binding, err)
		return nil, safeGatewayError(err)
	}
	data := call.Data
	if data == nil {
		data = map[string]any{}
	}
	if s.validator != nil && len(resolved.OutputSchema) > 0 {
		if err := s.validator.Validate(data, resolved.OutputSchema); err != nil {
			validationErr := domaincap.NewCapabilityError(domaincap.ErrorKindValidation, "capability.output_invalid", "capability provider returned an invalid result")
			s.recordFailedEvidence(ctx, req, projectID, stateID, resolved, binding, validationErr)
			return nil, validationErr
		}
	}
	rawResult, err := json.Marshal(data)
	if err != nil {
		return nil, domaincap.NewCapabilityError(domaincap.ErrorKindInternal, "capability.result_invalid", "capability result could not be normalized")
	}
	correlation := req.CorrelationID
	if _, err := s.evidence.Upsert(ctx, repositories.CapabilityEvidenceInput{
		TenantID: req.TenantID, ProjectID: projectID, WorkflowInstanceID: inst.ID, StateID: stateID,
		CapabilityID: resolved.ID, CapabilityName: resolved.Name, ProviderServer: binding.ConnectionAlias,
		ProviderTool: binding.ToolName, CorrelationID: &correlation, IdempotencyKey: req.IdempotencyKey,
		Status: entities.CapabilityEvidenceSucceeded, Result: rawResult,
	}); err != nil {
		return nil, gatewayUnavailable("capability evidence could not be persisted")
	}
	s.persistContext(ctx, req.TenantID, inst.ID, resolved.Name, data)
	return &GatewayInvocationResult{
		InstanceID: inst.ID, StateID: stateID, CapabilityName: resolved.Name,
		Data: data, Status: entities.CapabilityEvidenceSucceeded,
	}, nil
}

func (s *MCPGatewayService) lockIdempotency(parts ...string) func() {
	key := strings.Join(parts, "\x00")
	s.locksMu.Lock()
	if s.locks == nil {
		s.locks = map[string]*sync.Mutex{}
	}
	lock, ok := s.locks[key]
	if !ok {
		lock = &sync.Mutex{}
		s.locks[key] = lock
	}
	s.locksMu.Unlock()
	lock.Lock()
	return lock.Unlock
}

func (s *MCPGatewayService) executeNonMCP(ctx context.Context, req GatewayInvocationRequest, inst *entities.WorkflowInstance, stateID string, resolved *domaincap.ResolvedCapability) (*GatewayInvocationResult, error) {
	if s.fallback == nil {
		return nil, gatewayUnavailable("non-MCP capability gateway is not configured")
	}
	result, err := s.fallback.Execute(ctx, domaincap.Invocation{
		TenantID: req.TenantID, WorkflowID: inst.WorkflowID, WorkflowInstanceID: inst.ID,
		StateID: stateID, ActionID: req.IdempotencyKey, CapabilityID: resolved.ID,
		Name: resolved.Name, Payload: req.Payload, IdempotencyKey: req.IdempotencyKey,
		Policy: domaincap.CapabilityPolicy{Timeout: s.timeout},
	})
	if err != nil {
		return nil, safeGatewayError(err)
	}
	s.persistContext(ctx, req.TenantID, inst.ID, resolved.Name, result.Data)
	return &GatewayInvocationResult{InstanceID: inst.ID, StateID: stateID, CapabilityName: resolved.Name, Data: result.Data, Status: entities.CapabilityEvidenceSucceeded}, nil
}

func (s *MCPGatewayService) recordFailedEvidence(ctx context.Context, req GatewayInvocationRequest, projectID, stateID string, resolved *domaincap.ResolvedCapability, binding *entities.ProjectCapabilityMCPBinding, err error) {
	if s.evidence == nil || binding == nil || resolved == nil {
		return
	}
	ce := safeCapabilityError(err)
	raw, marshalErr := json.Marshal(map[string]string{"kind": string(ce.Kind), "code": ce.Code, "message": ce.Message})
	if marshalErr != nil {
		return
	}
	correlation := req.CorrelationID
	_, _ = s.evidence.Upsert(ctx, repositories.CapabilityEvidenceInput{
		TenantID: req.TenantID, ProjectID: projectID, WorkflowInstanceID: req.InstanceID, StateID: stateID,
		CapabilityID: resolved.ID, CapabilityName: resolved.Name, ProviderServer: binding.ConnectionAlias,
		ProviderTool: binding.ToolName, CorrelationID: &correlation, IdempotencyKey: req.IdempotencyKey,
		Status: entities.CapabilityEvidenceFailed, Error: raw,
	})
}

func (s *MCPGatewayService) persistContext(ctx context.Context, tenantID, instanceID, capabilityName string, data map[string]any) {
	if s.context == nil {
		return
	}
	for key, value := range data {
		if strings.HasPrefix(key, "_") {
			continue
		}
		raw, err := json.Marshal(value)
		if err != nil {
			continue
		}
		_, _ = s.context.UpsertContext(ctx, tenantID, entities.ContextScopeWorkflowInstance, instanceID, capabilityName+"."+key, raw, 0)
	}
}

func findBoundTool(tools []entities.MCPDiscoveredTool, binding *entities.ProjectCapabilityMCPBinding) *entities.MCPDiscoveredTool {
	for i := range tools {
		if tools[i].ID == binding.ToolID {
			return &tools[i]
		}
	}
	return nil
}

func declaredCapability(capabilities []string, wanted string) bool {
	for _, name := range capabilities {
		if strings.TrimSpace(name) == strings.TrimSpace(wanted) {
			return true
		}
	}
	return false
}

func decodeEvidenceResult(raw []byte) map[string]any {
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil || data == nil {
		return map[string]any{}
	}
	return data
}

func safeCapabilityError(err error) *domaincap.CapabilityError {
	var ce *domaincap.CapabilityError
	if errors.As(err, &ce) {
		return domaincap.NewCapabilityError(ce.Kind, safeGatewayErrorCode(ce.Kind), safeGatewayErrorMessage(ce.Kind))
	}
	return domaincap.NewCapabilityError(domaincap.ErrorKindExternal, "capability.failed", "MCP capability provider failed")
}

func safeGatewayErrorCode(kind domaincap.ErrorKind) string {
	switch kind {
	case domaincap.ErrorKindTimeout:
		return "capability.timeout"
	case domaincap.ErrorKindUnauthorized:
		return "capability.unauthorized"
	case domaincap.ErrorKindValidation:
		return "capability.validation_failed"
	case domaincap.ErrorKindRateLimited:
		return "capability.rate_limited"
	case domaincap.ErrorKindUnavailable:
		return "capability.unavailable"
	case domaincap.ErrorKindBusiness:
		return "capability.business_error"
	default:
		return "capability.failed"
	}
}

func safeGatewayErrorMessage(kind domaincap.ErrorKind) string {
	switch kind {
	case domaincap.ErrorKindTimeout:
		return "MCP capability invocation timed out"
	case domaincap.ErrorKindUnauthorized:
		return "capability is not authorized"
	case domaincap.ErrorKindValidation:
		return "capability input or output is invalid"
	case domaincap.ErrorKindRateLimited:
		return "capability rate limit exceeded"
	case domaincap.ErrorKindUnavailable:
		return "MCP capability provider is unavailable"
	case domaincap.ErrorKindBusiness:
		return "MCP provider rejected the capability request"
	default:
		return "MCP capability provider failed"
	}
}

func safeGatewayError(err error) error {
	ce := safeCapabilityError(err)
	return domaincap.NewCapabilityError(ce.Kind, ce.Code, ce.Message)
}

func gatewayUnavailable(message string) error {
	return gatewayFailure("capability.gateway_unavailable", message)
}

func gatewayFailure(code, message string) error {
	return domaincap.NewCapabilityError(domaincap.ErrorKindUnavailable, code, message)
}

func gatewayUnauthorized(message string) error {
	return domaincap.NewCapabilityError(domaincap.ErrorKindUnauthorized, "capability.gateway_unauthorized", message)
}

func isNotFound(err error) bool {
	var de *domain.DomainError
	return errors.As(err, &de) && de.Code == domain.ErrNotFound || errors.Is(err, sql.ErrNoRows)
}
