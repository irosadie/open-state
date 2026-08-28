package engine

import (
	"context"

	"github.com/irosadie/open-state/go-shared/domain"
)

// Replay re-drives a workflow instance's recorded events through a fresh,
// in-memory engine to reproduce the resulting context and current state without
// persisting anything (PRD 52, 170). It is deterministic and side-effect free.
func (e *Engine) Replay(ctx context.Context, tenantID, instanceID string, events []Event) (*WorkflowInstance, error) {
	orig, err := e.repos.Instances.Get(ctx, tenantID, instanceID)
	if err != nil {
		return nil, err
	}
	def, err := e.loadDefinition(ctx, tenantID, orig.ProjectID, orig.WorkflowID)
	if err != nil {
		return nil, err
	}

	// Fresh in-memory repos so replay never writes to the real store.
	repos := replayRepos()
	replayEngine := NewEngine(repos)
	if err := repos.Workflows.Save(ctx, def); err != nil {
		return nil, err
	}

	// Start a fresh instance at the entry node.
	inst, err := replayEngine.StartWorkflow(ctx, tenantID, orig.ProjectID, orig.ConversationID, def, entryEvent(def))
	if err != nil {
		return nil, err
	}

	// Re-drive each recorded event in deterministic sequence order.
	for i := range events {
		evt := &events[i]
		next, _, err := replayEngine.ProcessEvent(ctx, tenantID, inst.ID, &Event{
			ID:          evt.ID,
			Type:        evt.Type,
			Source:      evt.Source,
			Payload:     evt.Payload,
			Timestamp:   evt.Timestamp,
			IdempotencyKey: evt.IdempotencyKey,
		})
		if err != nil {
			return nil, err
		}
		inst = next
	}
	return inst, nil
}

// entryEvent picks the trigger event used to start a fresh instance, preferring the
// first workflow trigger, else a default "workflow.started".
func entryEvent(def *WorkflowDefinition) string {
	if len(def.Triggers) > 0 && def.Triggers[0].Event != "" {
		return def.Triggers[0].Event
	}
	return "workflow.started"
}

// replayRepos builds in-memory repository ports for a non-persisting replay.
func replayRepos() EngineRepositories {
	return EngineRepositories{
		Projects:  replayProjectRepo{},
		Workflows: &replayWorkflowRepo{defs: map[string]*WorkflowDefinition{}},
		Instances: &replayInstanceRepo{insts: map[string]*WorkflowInstance{}},
		Events:    &replayEventRepo{processed: map[string]bool{}},
	}
}

type replayProjectRepo struct{}

func (replayProjectRepo) Get(context.Context, string, string) (*Project, error) {
	return &Project{ID: "proj"}, nil
}
func (replayProjectRepo) Save(context.Context, *Project) error { return nil }

type replayWorkflowRepo struct{ defs map[string]*WorkflowDefinition }

func (r *replayWorkflowRepo) GetBySlug(_ context.Context, _, projectID, slug string) (*WorkflowDefinition, error) {
	d, ok := r.defs[projectID+"/"+slug]
	if !ok {
		return nil, domain.NewNotFound("workflow not found")
	}
	return d, nil
}
func (r *replayWorkflowRepo) Save(_ context.Context, d *WorkflowDefinition) error {
	r.defs[d.ProjectID+"/"+d.Slug] = d
	return nil
}

type replayInstanceRepo struct{ insts map[string]*WorkflowInstance }

func (r *replayInstanceRepo) Create(_ context.Context, i *WorkflowInstance) error {
	cp := *i
	r.insts[i.ID] = &cp
	return nil
}
func (r *replayInstanceRepo) Get(_ context.Context, _, id string) (*WorkflowInstance, error) {
	i, ok := r.insts[id]
	if !ok {
		return nil, domain.NewNotFound("instance not found")
	}
	cp := *i
	return &cp, nil
}
func (r *replayInstanceRepo) UpdateWithVersion(_ context.Context, i *WorkflowInstance, expected int) error {
	cur, ok := r.insts[i.ID]
	if !ok {
		return domain.NewNotFound("instance not found")
	}
	if cur.Version != expected {
		return domain.NewConflict("version conflict")
	}
	cp := *i
	r.insts[i.ID] = &cp
	return nil
}

type replayEventRepo struct{ processed map[string]bool }

func (replayEventRepo) Append(context.Context, *Event) error { return nil }
func (r *replayEventRepo) IsProcessed(_ context.Context, _, key string) (bool, error) {
	return r.processed[key], nil
}
func (r *replayEventRepo) MarkProcessed(_ context.Context, _, key, _ string) error {
	r.processed[key] = true
	return nil
}
