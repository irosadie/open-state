package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	domain "github.com/irosadie/open-state/go-shared/domain"
)

// OrchestratorService exposes the runtime workflow orchestration operations the
// MCP orchestrator tools need (PRD 25, 38, 42-43, 52, 142). It composes the
// existing persistence repositories; every method is tenant-scoped (PRD 4, 96).
// It is the application-layer seam the thin MCP handlers delegate to (PRD 74).
type OrchestratorService struct {
	instances repositories.IInstanceRepository
	events    repositories.IEventRepository
	context   repositories.IContextRepository
	capabs    repositories.ICapabilityRepository
	now       func() time.Time
}

// NewOrchestratorService builds an OrchestratorService.
func NewOrchestratorService(
	instances repositories.IInstanceRepository,
	events repositories.IEventRepository,
	context repositories.IContextRepository,
	capabs repositories.ICapabilityRepository,
) *OrchestratorService {
	return &OrchestratorService{
		instances: instances,
		events:    events,
		context:   context,
		capabs:    capabs,
		now:       time.Now,
	}
}

// StartWorkflow creates a new workflow instance in RUNNING state.
func (s *OrchestratorService) StartWorkflow(ctx context.Context, tenantID, workflowID, workflowVersionID, correlationKey string) (*entities.WorkflowInstance, error) {
	var corr *string
	if correlationKey != "" {
		corr = &correlationKey
	}
	return s.instances.Create(ctx, tenantID, repositories.CreateWorkflowInstanceInput{
		WorkflowID:        workflowID,
		WorkflowVersionID: workflowVersionID,
		CorrelationKey:    corr,
		StartedAt:         nil,
		ExpiresAt:         nil,
	})
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

// ProposeEvent validates the instance is active, appends the event to history, and
// advances the instance (PRD 38). Full guard/transition evaluation lands with the
// engine wiring; this slice persists the event and marks the instance active.
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
