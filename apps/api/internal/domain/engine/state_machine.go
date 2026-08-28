package engine

import (
	"context"
	"errors"
	"time"

	"github.com/irosadie/open-state/go-shared/domain"
)

// Engine executes workflow definitions deterministically.
// It is domain-pure and depends only on repository ports.
type Engine struct {
	repos EngineRepositories
	now   func() time.Time
}

// NewEngine constructs the engine with the provided repository ports.
func NewEngine(repos EngineRepositories) *Engine {
	return &Engine{repos: repos, now: time.Now}
}

// StartWorkflow creates a workflow instance and enters the START node.
// The workflow is resolved within the given project (hierarchy: project → workflow).
func (e *Engine) StartWorkflow(ctx context.Context, tenantID, projectID, conversationID string, def *WorkflowDefinition, entryEvent string) (*WorkflowInstance, error) {
	if def == nil || def.EntryNodeID == "" {
		return nil, domain.NewValidation("workflow definition or entry node is missing")
	}
	if len(def.Nodes) == 0 {
		return nil, domain.NewValidation("workflow has no nodes")
	}

	now := e.now()
	instance := &WorkflowInstance{
		ID:                newID(),
		TenantID:          tenantID,
		ProjectID:         projectID,
		WorkflowID:        def.Slug,
		WorkflowVersionID: defSlugVersion(def),
		ConversationID:    conversationID,
		Status:            InstanceRunning,
		CurrentStateID:    def.EntryNodeID,
		Context:           map[string]any{},
		Version:           1,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if err := e.repos.Instances.Create(ctx, instance); err != nil {
		return nil, err
	}
	return instance, nil
}

// ProcessEvent runs the deterministic event pipeline:
// load → validate event allowed → evaluate guards → pick transition →
// apply → emit result. (PRD §152)
func (e *Engine) ProcessEvent(ctx context.Context, tenantID, instanceID string, evt *Event) (*WorkflowInstance, *StateInstance, error) {
	if evt == nil {
		return nil, nil, domain.NewValidation("event is required")
	}
	instance, err := e.repos.Instances.Get(ctx, tenantID, instanceID)
	if err != nil {
		return nil, nil, err
	}
	if instance.Status != InstanceRunning && instance.Status != InstanceWaiting {
		return nil, nil, domain.NewConflict("workflow instance is not active (status: " + string(instance.Status) + ")")
	}

	// idempotency: skip if already processed
	if evt.IdempotencyKey != "" {
		done, err := e.repos.Events.IsProcessed(ctx, tenantID, evt.IdempotencyKey)
		if err != nil {
			return nil, nil, err
		}
		if done {
			return instance, nil, nil // no-op, deduped (PRD §30)
		}
	}

	def, err := e.loadDefinition(ctx, tenantID, instance.ProjectID, instance.WorkflowID)
	if err != nil {
		return nil, nil, err
	}

	if _, ok := nodeByID(def, instance.CurrentStateID); !ok {
		return nil, nil, domain.NewValidation("current state not found in workflow")
	}

	// validate event allowed from current state
	candidates := transitionsFrom(def, instance.CurrentStateID, evt.Type)
	if len(candidates) == 0 {
		return nil, nil, domain.NewConflict("event not allowed from current state: " + evt.Type)
	}

	// merge event payload into context BEFORE guard evaluation so guards
	// can inspect data carried by the event (e.g. slot.available).
	if evt.Payload != nil {
		if instance.Context == nil {
			instance.Context = map[string]any{}
		}
		mergePayload(instance.Context, evt.Payload)
	}

	// evaluate guards and pick highest-priority passing transition (PRD §34)
	transition, err := selectTransition(candidates, instance.Context)
	if err != nil {
		return nil, nil, err
	}
	if transition == nil {
		return nil, nil, &ErrGuardFailed{Message: "no candidate transition passed guards"}
	}

	// apply transition
	target, ok := nodeByID(def, transition.TargetStateID)
	if !ok {
		return nil, nil, domain.NewValidation("transition target not found: " + transition.TargetStateID)
	}

	now := e.now()
	instance.CurrentStateID = target.ID
	instance.Status = InstanceRunning
	instance.Version++

	if err := e.repos.Instances.UpdateWithVersion(ctx, instance, instance.Version-1); err != nil {
		return nil, nil, err
	}

	// append event + mark idempotency
	if err := e.repos.Events.Append(ctx, evt); err != nil {
		return nil, nil, err
	}
	if evt.IdempotencyKey != "" {
		if err := e.repos.Events.MarkProcessed(ctx, tenantID, evt.IdempotencyKey, evt.ID); err != nil {
			return nil, nil, err
		}
	}

	stateInst := &StateInstance{
		ID:                newID(),
		WorkflowInstanceID: instance.ID,
		StateID:           target.ID,
		Status:            StateActive,
		EnteredAt:         now,
		ExitedAt:          ptrTime(now),
	}

	return instance, stateInst, nil
}

func (e *Engine) loadDefinition(ctx context.Context, tenantID, projectID, slug string) (*WorkflowDefinition, error) {
	def, err := e.repos.Workflows.GetBySlug(ctx, tenantID, projectID, slug)
	if err != nil {
		return nil, err
	}
	return def, nil
}

// AllowedTransitions returns the transitions available from the instance's current
// state (derived from its pinned workflow definition), so a client knows which
// events it may propose next (PRD 12, 14, 33-34).
func (e *Engine) AllowedTransitions(ctx context.Context, tenantID, instanceID string) ([]TransitionDefinition, error) {
	instance, err := e.repos.Instances.Get(ctx, tenantID, instanceID)
	if err != nil {
		return nil, err
	}
	def, err := e.loadDefinition(ctx, tenantID, instance.ProjectID, instance.WorkflowID)
	if err != nil {
		return nil, err
	}
	var out []TransitionDefinition
	for _, t := range def.Transitions {
		if t.SourceStateID == instance.CurrentStateID {
			out = append(out, t)
		}
	}
	return out, nil
}

func nodeByID(def *WorkflowDefinition, id string) (WorkflowNode, bool) {
	for _, n := range def.Nodes {
		if n.ID == id {
			return n, true
		}
	}
	return WorkflowNode{}, false
}

func transitionsFrom(def *WorkflowDefinition, sourceStateID, eventType string) []TransitionDefinition {
	var out []TransitionDefinition
	for _, t := range def.Transitions {
		if t.SourceStateID == sourceStateID && t.Event == eventType {
			out = append(out, t)
		}
	}
	return out
}

// selectTransition returns the highest-priority transition whose guards pass.
// Returns nil if none pass. (PRD §34 — lower numeric priority evaluated first)
//
// A candidate whose guards fail (ErrGuardFailed) is skipped, not fatal, so the
// highest-priority *passing* transition wins. Only genuine evaluation errors
// (e.g. an unsupported operator) abort the selection.
func selectTransition(candidates []TransitionDefinition, ctx map[string]any) (*TransitionDefinition, error) {
	var best *TransitionDefinition
	for i := range candidates {
		c := candidates[i]
		ok, err := EvaluateGuards(c.Guards, ctx)
		if err != nil {
			var guardErr *ErrGuardFailed
			if errors.As(err, &guardErr) {
				continue // this candidate doesn't apply; try the next
			}
			return nil, err
		}
		if !ok {
			continue
		}
		if best == nil || c.Priority < best.Priority {
			cc := c
			best = &cc
		}
	}
	return best, nil
}

func mergePayload(ctx map[string]any, payload map[string]any) {
	for k, v := range payload {
		ctx[k] = v
	}
}

func ptrTime(t time.Time) *time.Time { return &t }
