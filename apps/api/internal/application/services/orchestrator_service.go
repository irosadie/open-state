package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/irosadie/open-state/api/internal/domain/engine"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	domain "github.com/irosadie/open-state/go-shared/domain"
)

// OrchestratorService exposes the runtime workflow orchestration operations the
// MCP orchestrator tools need (PRD 25, 38, 42-43, 52, 142). It composes the
// existing persistence repositories; every method is tenant-scoped (PRD 4, 96).
// It is the application-layer seam the thin MCP handlers delegate to (PRD 74).
//
// An optional engine may be injected to back propose/current-state/replay with
// real state-machine evaluation (PRD 170). When the engine is nil, the service
// degrades to the append/merge behavior so non-engine callers and tests keep
// working.
type OrchestratorService struct {
	instances   repositories.IInstanceRepository
	events      repositories.IEventRepository
	context     repositories.IContextRepository
	capabs      repositories.ICapabilityRepository
	workflows   repositories.IWorkflowRepository
	evidence    repositories.ICapabilityEvidenceRepository
	mcpBindings repositories.IProjectCapabilityMCPBindingRepository
	engine      *engine.Engine
	now         func() time.Time
}

// SetProjectCapabilityMCPBindings wires the project-scoped binding repository
// into the transition gate without changing the compatibility constructors.
// When configured, active bindings—not legacy capability provider fields—are
// required for external MCP evidence to satisfy a state.
func (s *OrchestratorService) SetProjectCapabilityMCPBindings(repo repositories.IProjectCapabilityMCPBindingRepository) {
	s.mcpBindings = repo
}

// NewOrchestratorService builds an OrchestratorService without an engine.
func NewOrchestratorService(
	instances repositories.IInstanceRepository,
	events repositories.IEventRepository,
	context repositories.IContextRepository,
	capabs repositories.ICapabilityRepository,
	workflows repositories.IWorkflowRepository,
	evidence ...repositories.ICapabilityEvidenceRepository,
) *OrchestratorService {
	svc := &OrchestratorService{
		instances: instances,
		events:    events,
		context:   context,
		capabs:    capabs,
		workflows: workflows,
		now:       time.Now,
	}
	if len(evidence) > 0 {
		svc.evidence = evidence[0]
	}
	return svc
}

// NewEngineOrchestratorService builds an OrchestratorService wired to a runtime
// engine, so propose/current-state/replay execute real state transitions.
func NewEngineOrchestratorService(
	instances repositories.IInstanceRepository,
	events repositories.IEventRepository,
	context repositories.IContextRepository,
	capabs repositories.ICapabilityRepository,
	eng *engine.Engine,
	workflows repositories.IWorkflowRepository,
	evidence ...repositories.ICapabilityEvidenceRepository,
) *OrchestratorService {
	svc := &OrchestratorService{
		instances: instances,
		events:    events,
		context:   context,
		capabs:    capabs,
		workflows: workflows,
		engine:    eng,
		now:       time.Now,
	}
	if len(evidence) > 0 {
		svc.evidence = evidence[0]
	}
	return svc
}

// StartWorkflow creates a new workflow instance in RUNNING state.
// If workflowVersionID is empty, the current published version is resolved automatically.
func (s *OrchestratorService) StartWorkflow(ctx context.Context, tenantID, workflowID, workflowVersionID, correlationKey string) (*entities.WorkflowInstance, error) {
	if workflowVersionID == "" {
		if s.workflows == nil {
			return nil, domain.NewValidation("workflow_version_id is required (no workflow repository configured)")
		}
		ver, err := s.workflows.FindCurrentVersionByWorkflow(ctx, tenantID, workflowID)
		if err != nil {
			return nil, err
		}
		workflowVersionID = ver.ID
	}
	var corr *string
	if correlationKey != "" {
		corr = &correlationKey
	}
	created, err := s.instances.Create(ctx, tenantID, repositories.CreateWorkflowInstanceInput{
		WorkflowID:        workflowID,
		WorkflowVersionID: workflowVersionID,
		CorrelationKey:    corr,
		StartedAt:         nil,
		ExpiresAt:         nil,
	})
	if err != nil {
		return nil, err
	}
	if s.engine == nil || created.Status != entities.WorkflowInstanceCreated {
		return created, nil
	}

	started, err := s.engine.InitializeWorkflow(ctx, tenantID, created.ID)
	if err != nil {
		return nil, err
	}
	return s.instances.UpdateStatus(ctx, tenantID, created.ID, entities.WorkflowInstanceRunning, started.Version)
}

// SuspendWorkflow pauses a running instance (PRD 42-43).
func (s *OrchestratorService) SuspendWorkflow(ctx context.Context, tenantID, instanceID string) (*entities.WorkflowInstance, error) {
	inst, err := s.instances.FindByID(ctx, tenantID, instanceID)
	if err != nil {
		return nil, err
	}
	if inst.Status != entities.WorkflowInstanceRunning && inst.Status != entities.WorkflowInstanceWaiting {
		return nil, domain.NewConflict("only running/waiting instances can be suspended")
	}
	return s.instances.UpdateStatus(ctx, tenantID, instanceID, entities.WorkflowInstanceSuspended, inst.Version)
}

// ResumeWorkflow restores a suspended instance to RUNNING (PRD 42-43).
func (s *OrchestratorService) ResumeWorkflow(ctx context.Context, tenantID, instanceID string) (*entities.WorkflowInstance, error) {
	inst, err := s.instances.FindByID(ctx, tenantID, instanceID)
	if err != nil {
		return nil, err
	}
	if inst.Status != entities.WorkflowInstanceSuspended {
		return nil, domain.NewConflict("only suspended instances can be resumed")
	}
	return s.instances.UpdateStatus(ctx, tenantID, instanceID, entities.WorkflowInstanceRunning, inst.Version)
}

// RetryWorkflow retries the current failed state using its optimistic version.
// It never overwrites a newer state projection and only accepts failed runtime
// state, leaving event history immutable.
func (s *OrchestratorService) RetryWorkflow(ctx context.Context, tenantID, instanceID string) (*entities.WorkflowInstance, error) {
	inst, state, err := s.GetCurrentState(ctx, tenantID, instanceID)
	if err != nil {
		return nil, err
	}
	if state == nil || (state.Status != entities.StateInstanceFailed && inst.Status != entities.WorkflowInstanceFailed) {
		return nil, domain.NewConflict("only failed instances can be retried")
	}

	retriedState, err := s.instances.IncrementRetry(ctx, tenantID, state.ID, state.Version)
	if err != nil {
		return nil, err
	}
	if retriedState == nil {
		return nil, domain.NewInternal("retry did not return the updated state")
	}
	if state.Status == entities.StateInstanceFailed {
		if _, err := s.instances.UpdateStateInstanceStatus(ctx, tenantID, state.ID, entities.StateInstanceActive, retriedState.Version); err != nil {
			return nil, err
		}
	}
	if inst.Status == entities.WorkflowInstanceFailed {
		return s.instances.UpdateStatus(ctx, tenantID, instanceID, entities.WorkflowInstanceRunning, inst.Version)
	}
	return inst, nil
}

// CancelWorkflow cancels a non-terminal instance (PRD 42-43).
func (s *OrchestratorService) CancelWorkflow(ctx context.Context, tenantID, instanceID string) (*entities.WorkflowInstance, error) {
	inst, err := s.instances.FindByID(ctx, tenantID, instanceID)
	if err != nil {
		return nil, err
	}
	if inst.Status == entities.WorkflowInstanceCompleted || inst.Status == entities.WorkflowInstanceCancelled {
		return nil, domain.NewConflict("workflow already terminal")
	}
	return s.instances.UpdateStatus(ctx, tenantID, instanceID, entities.WorkflowInstanceCancelled, inst.Version)
}

// ListInstances returns the tenant's workflow instances, newest first (PRD 142).
func (s *OrchestratorService) ListInstances(ctx context.Context, tenantID string) ([]entities.WorkflowInstance, error) {
	return s.instances.ListByTenant(ctx, tenantID)
}

// GetCurrentState returns the workflow instance and its current state instance id.
func (s *OrchestratorService) GetCurrentState(ctx context.Context, tenantID, instanceID string) (*entities.WorkflowInstance, *entities.StateInstance, error) {
	inst, err := s.instances.FindByID(ctx, tenantID, instanceID)
	if err != nil {
		return nil, nil, err
	}
	if inst.CurrentStateInstanceID == nil || *inst.CurrentStateInstanceID == "" {
		return inst, nil, nil
	}
	stateInst, err := s.instances.FindStateInstanceByID(ctx, tenantID, *inst.CurrentStateInstanceID)
	if err != nil {
		return inst, nil, nil // state projection may be absent; return instance alone
	}
	return inst, stateInst, nil
}

// ListHistory returns the event history for an instance in deterministic sequence
// order (PRD 52).
func (s *OrchestratorService) ListHistory(ctx context.Context, tenantID, instanceID string) ([]entities.Event, error) {
	return s.events.ListEventsByInstance(ctx, tenantID, instanceID)
}

// GetAllowedTransitions returns the transitions available from the instance's
// current state. When an engine is wired, it derives them from the pinned workflow
// definition; otherwise it returns an empty list.
func (s *OrchestratorService) GetAllowedTransitions(ctx context.Context, tenantID, instanceID string) ([]engine.TransitionDefinition, error) {
	if s.engine == nil {
		return []engine.TransitionDefinition{}, nil
	}
	return s.engine.AllowedTransitions(ctx, tenantID, instanceID)
}

// CurrentStateInfo returns the current node's purpose/instructions/context when an
// engine is wired; otherwise an empty StateInfo.
func (s *OrchestratorService) CurrentStateInfo(ctx context.Context, tenantID, instanceID string) (*engine.StateInfo, error) {
	if s.engine == nil {
		return &engine.StateInfo{StateID: instanceID}, nil
	}
	return s.engine.CurrentStateInfo(ctx, tenantID, instanceID)
}

// ReplayState replays the recorded events through the engine to reproduce the
// resulting context and current state (PRD 52). When no engine is wired, it falls
// back to the merge-based context snapshot with an empty state key.
func (s *OrchestratorService) ReplayState(ctx context.Context, tenantID, instanceID string) (map[string]any, string, error) {
	if s.engine == nil {
		snap, _, err := s.ReplayWorkflow(ctx, tenantID, instanceID)
		return snap, "", err
	}
	events, err := s.events.ListEventsByInstance(ctx, tenantID, instanceID)
	if err != nil {
		return nil, "", err
	}
	engineEvents := make([]engine.Event, 0, len(events))
	for i := range events {
		e := &events[i]
		var payload map[string]any
		if len(e.Payload) > 0 {
			_ = json.Unmarshal(e.Payload, &payload)
		}
		engineEvents = append(engineEvents, engine.Event{
			ID:                 e.EventID,
			TenantID:           tenantID,
			Type:               e.Type,
			Source:             engine.EventSource(e.Source),
			WorkflowInstanceID: instanceID,
			Payload:            payload,
			Timestamp:          e.Timestamp,
		})
	}
	replayed, err := s.engine.Replay(ctx, tenantID, instanceID, engineEvents)
	if err != nil {
		return nil, "", err
	}
	return replayed.Context, replayed.CurrentStateID, nil
}

// ProposeEvent validates the instance is active and, when an engine is wired, runs
// the engine's `event → guard → transition` evaluation and persists the resulting
// state (PRD 38, §34). Without an engine, it falls back to appending the event to
// history and marking the instance active.
func (s *OrchestratorService) ProposeEvent(ctx context.Context, tenantID, instanceID, eventType string, payload map[string]any, correlationID string) (*entities.Event, error) {
	if eventType == "" {
		return nil, domain.NewValidation("event type is required")
	}
	inst, err := s.instances.FindByID(ctx, tenantID, instanceID)
	if err != nil {
		return nil, err
	}
	if inst.Status != entities.WorkflowInstanceRunning && inst.Status != entities.WorkflowInstanceWaiting {
		return nil, domain.NewConflict("workflow instance is not active")
	}

	if s.engine != nil {
		if err := s.validateCapabilityEvidence(ctx, tenantID, inst, instanceID); err != nil {
			return nil, err
		}
		evt := &engine.Event{
			ID:                 uuid.NewString(),
			TenantID:           tenantID,
			Type:               eventType,
			Source:             engine.SourceMCP,
			WorkflowInstanceID: instanceID,
			Payload:            payload,
			Timestamp:          time.Now().UTC(),
		}
		// The engine loads the instance, evaluates guards, transitions, and persists.
		next, _, err := s.engine.ProcessEvent(ctx, tenantID, instanceID, evt)
		if err != nil {
			return nil, err
		}
		_ = next
		return &entities.Event{
			ID:                 evt.ID,
			EventID:            evt.ID,
			TenantID:           tenantID,
			Type:               eventType,
			Source:             entities.EventSourceMCP,
			WorkflowInstanceID: &instanceID,
			Timestamp:          evt.Timestamp,
		}, nil
	}

	raw := []byte("{}")
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return nil, domain.NewValidation("invalid event payload")
		}
		raw = b
	}
	var corr *string
	if correlationID != "" {
		corr = &correlationID
	}

	now := time.Now().UTC()
	evt := repositories.AppendEventInput{
		EventID:            uuid.NewString(),
		Type:               eventType,
		Source:             entities.EventSourceMCP,
		AggregateID:        nil,
		WorkflowInstanceID: &instanceID,
		Timestamp:          now,
		Payload:            raw,
		CorrelationID:      corr,
		CausationID:        nil,
		IdempotencyKey:     nil,
	}
	return s.events.Append(ctx, tenantID, evt)
}

// validateCapabilityEvidence enforces the State MCP gate for external MCP
// capabilities. It runs before engine.ProcessEvent so a rejected proposal
// cannot mutate the instance, context, or event history.
func (s *OrchestratorService) validateCapabilityEvidence(ctx context.Context, tenantID string, inst *entities.WorkflowInstance, instanceID string) error {
	if s.evidence == nil || s.capabs == nil {
		return nil
	}
	info, err := s.engine.CurrentStateInfo(ctx, tenantID, instanceID)
	if err != nil {
		return err
	}
	if len(info.Capabilities) == 0 {
		return nil
	}
	evidence, err := s.evidence.ListByInstanceState(ctx, tenantID, instanceID, info.StateID)
	if err != nil {
		return err
	}
	satisfied := make(map[string]bool, len(evidence))
	for _, item := range evidence {
		if item.Status == entities.CapabilityEvidenceSucceeded {
			satisfied[item.CapabilityID] = true
		}
	}
	for _, name := range info.Capabilities {
		cap, findErr := s.capabs.FindByName(ctx, tenantID, name)
		if findErr != nil {
			return domain.NewConflict("capability requirement not satisfied: " + name)
		}
		// Internal capabilities continue through invoke_capability. This gate is
		// specifically for the direct two-MCP provider path.
		if cap.ProviderType != entities.ProviderTypeMCP {
			continue
		}
		if s.mcpBindings != nil {
			projectID := info.ProjectID
			if projectID == "" && s.workflows != nil {
				if version, versionErr := s.workflows.FindCurrentVersionByWorkflow(ctx, tenantID, inst.WorkflowID); versionErr == nil {
					projectID = version.ProjectID
				}
			}
			binding, bindingErr := s.mcpBindings.FindByCapability(ctx, tenantID, projectID, cap.ID)
			if bindingErr != nil || binding == nil || binding.Health != entities.ProjectCapabilityMCPBindingActive || !satisfied[cap.ID] {
				return domain.NewConflict("capability requirement not satisfied: " + name)
			}
		} else if !cap.ProviderID.Valid || !cap.ProviderTool.Valid || !satisfied[cap.ID] {
			return domain.NewConflict("capability requirement not satisfied: " + name)
		}
	}
	_ = inst
	return nil
}

// ListAllowedCapabilities returns the capabilities authorized for a scope (PRD 59-62).
func (s *OrchestratorService) ListAllowedCapabilities(ctx context.Context, tenantID string, scopeType entities.BindingScopeType, scopeID string) ([]entities.Capability, error) {
	bindings, err := s.capabs.ListBindingsByScope(ctx, tenantID, scopeType, scopeID)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var out []entities.Capability
	for _, b := range bindings {
		if b.Permission != entities.BindingPermissionAllow || seen[b.CapabilityID] {
			continue
		}
		cap, err := s.capabs.FindByID(ctx, tenantID, b.CapabilityID)
		if err != nil {
			continue
		}
		seen[cap.ID] = true
		out = append(out, *cap)
	}
	return out, nil
}

// GetActiveWorkflow resolves the active (non-terminal) workflow instance for a
// conversation (PRD 10, 142). A conversation maps to a workflow instance via its
// correlation_key (PRD 6).
func (s *OrchestratorService) GetActiveWorkflow(ctx context.Context, tenantID, conversationID string) (*entities.WorkflowInstance, error) {
	instances, err := s.instances.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	for i := range instances {
		inst := &instances[i]
		if !inst.CorrelationKey.Valid || inst.CorrelationKey.String != conversationID {
			continue
		}
		if inst.Status == entities.WorkflowInstanceCompleted ||
			inst.Status == entities.WorkflowInstanceCancelled ||
			inst.Status == entities.WorkflowInstanceAborted ||
			inst.Status == entities.WorkflowInstanceExpired {
			continue
		}
		return inst, nil
	}
	return nil, domain.NewNotFound("no active workflow for conversation")
}

// ReplayWorkflow replays the recorded event history of an instance to reproduce its
// resulting context/state projection (PRD 52). It verifies the instance exists, then
// merges event payloads in sequence order and returns the reproduced context snapshot
// plus the last event. (Full engine re-execution of the replay is a refinement; the
// merge approach yields the deterministic context + last event the handler needs.)
func (s *OrchestratorService) ReplayWorkflow(ctx context.Context, tenantID, instanceID string) (map[string]any, *entities.Event, error) {
	if _, err := s.instances.FindByID(ctx, tenantID, instanceID); err != nil {
		return nil, nil, err
	}
	events, err := s.events.ListEventsByInstance(ctx, tenantID, instanceID)
	if err != nil {
		return nil, nil, err
	}

	contextSnap := map[string]any{}
	var last *entities.Event
	for i := range events {
		e := &events[i]
		if len(e.Payload) > 0 {
			var payload map[string]any
			if err := json.Unmarshal(e.Payload, &payload); err == nil {
				for k, v := range payload {
					contextSnap[k] = v
				}
			}
		}
		last = e
	}
	return contextSnap, last, nil
}
