package engine

import (
	"context"
	"testing"

	"github.com/irosadie/open-state/go-shared/domain"
)

// ---------------------------------------------------------------------------
// In-memory fakes for the repository ports (proves engine is DB-agnostic).
// ---------------------------------------------------------------------------

type fakeWorkflowRepo struct {
	defs map[string]*WorkflowDefinition
}

func workflowKey(projectID, slug string) string { return projectID + "/" + slug }

func (f *fakeWorkflowRepo) GetBySlug(_ context.Context, _ string, projectID, slug string) (*WorkflowDefinition, error) {
	d, ok := f.defs[workflowKey(projectID, slug)]
	if !ok {
		return nil, domain.NewNotFound("workflow not found")
	}
	return d, nil
}
func (f *fakeWorkflowRepo) Save(_ context.Context, def *WorkflowDefinition) error {
	if f.defs == nil {
		f.defs = map[string]*WorkflowDefinition{}
	}
	f.defs[workflowKey(def.ProjectID, def.Slug)] = def
	return nil
}

type fakeInstanceRepo struct {
	insts map[string]*WorkflowInstance
}

func (f *fakeInstanceRepo) Create(_ context.Context, inst *WorkflowInstance) error {
	if f.insts == nil {
		f.insts = map[string]*WorkflowInstance{}
	}
	f.insts[inst.ID] = cloneInstance(inst)
	return nil
}
func (f *fakeInstanceRepo) Get(_ context.Context, _ string, id string) (*WorkflowInstance, error) {
	i, ok := f.insts[id]
	if !ok {
		return nil, domain.NewNotFound("instance not found")
	}
	return cloneInstance(i), nil
}
func (f *fakeInstanceRepo) UpdateWithVersion(_ context.Context, inst *WorkflowInstance, expected int) error {
	cur, ok := f.insts[inst.ID]
	if !ok {
		return domain.NewNotFound("instance not found")
	}
	if cur.Version != expected {
		return domain.NewConflict("version conflict")
	}
	f.insts[inst.ID] = cloneInstance(inst)
	return nil
}

// cloneInstance returns a copy so the engine cannot mutate the stored record
// before an optimistic update (mirrors real DB isolation).
func cloneInstance(i *WorkflowInstance) *WorkflowInstance {
	c := *i
	if i.Context != nil {
		ctxCopy := make(map[string]any, len(i.Context))
		for k, v := range i.Context {
			ctxCopy[k] = v
		}
		c.Context = ctxCopy
	}
	return &c
}

type fakeEventRepo struct {
	events    []*Event
	processed map[string]bool
}

func (f *fakeEventRepo) Append(_ context.Context, evt *Event) error {
	f.events = append(f.events, evt)
	return nil
}
func (f *fakeEventRepo) IsProcessed(_ context.Context, _ string, key string) (bool, error) {
	if f.processed == nil {
		return false, nil
	}
	return f.processed[key], nil
}
func (f *fakeEventRepo) MarkProcessed(_ context.Context, _ string, key, _ string) error {
	if f.processed == nil {
		f.processed = map[string]bool{}
	}
	f.processed[key] = true
	return nil
}

type fakeProjectRepo struct {
	projects map[string]*Project
}

func (f *fakeProjectRepo) Get(_ context.Context, _ string, projectID string) (*Project, error) {
	p, ok := f.projects[projectID]
	if !ok {
		return nil, domain.NewNotFound("project not found")
	}
	return p, nil
}
func (f *fakeProjectRepo) Save(_ context.Context, project *Project) error {
	if f.projects == nil {
		f.projects = map[string]*Project{}
	}
	f.projects[project.ID] = project
	return nil
}

func newFakeRepos() EngineRepositories {
	return EngineRepositories{
		Projects:  &fakeProjectRepo{},
		Workflows: &fakeWorkflowRepo{},
		Instances: &fakeInstanceRepo{},
		Events:    &fakeEventRepo{},
	}
}

// ---------------------------------------------------------------------------
// Test fixture: a minimal PADEL-like workflow.
// ---------------------------------------------------------------------------

func padelDef() *WorkflowDefinition {
	prio := 300
	return &WorkflowDefinition{
		Slug:          "padel-booking",
		ProjectID:     "project-padel",
		Name:          "Padel Booking",
		SchemaVersion: 1,
		Status:        WorkflowPublished,
		EntryNodeID:   "start",
		Nodes: []WorkflowNode{
			{ID: "start", Kind: NodeKindStart, Name: "START"},
			{ID: "select_time", Kind: NodeKindState, Name: "SELECT_TIME", RequiredContext: []string{"booking.date", "booking.time"}},
			{ID: "check_stock", Kind: NodeKindDecision, Name: "CHECK_STOCK", RequiredContext: []string{"slot.available"}},
			{ID: "confirm", Kind: NodeKindState, Name: "CONFIRM", RequiredContext: []string{"booking.id"}},
			{ID: "done", Kind: NodeKindEnd, Name: "DONE", IsTerminal: true},
		},
		Transitions: []TransitionDefinition{
			{ID: "t1", SourceStateID: "start", Event: "workflow.started", TargetStateID: "select_time", Priority: 1},
			{ID: "t2", SourceStateID: "select_time", Event: "datetime.selected", TargetStateID: "check_stock", Priority: 1},
			{ID: "t3", SourceStateID: "check_stock", Event: "slot.available", TargetStateID: "confirm", Priority: 1,
				Guards: []GuardGroup{{Logic: "AND", Conditions: []GuardCondition{{Field: "slot.available", Operator: OpEq, Value: true}}}}},
			{ID: "t4", SourceStateID: "check_stock", Event: "slot.unavailable", TargetStateID: "start", Priority: 2,
				Guards: []GuardGroup{{Logic: "AND", Conditions: []GuardCondition{{Field: "slot.available", Operator: OpEq, Value: false}}}}},
			{ID: "t5", SourceStateID: "confirm", Event: "confirm.requested", TargetStateID: "done", Priority: 1},
		},
		Policy: WorkflowPolicy{MaxDurationSeconds: &prio, Interruptible: "USER_REQUESTED", Priority: 10},
	}
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestStartWorkflow(t *testing.T) {
	repos := newFakeRepos()
	_ = repos.Workflows.Save(context.Background(), padelDef())
	eng := NewEngine(repos)

	inst, err := eng.StartWorkflow(context.Background(), "tenant1", "project-padel", "conv1", padelDef(), "workflow.started")
	if err != nil {
		t.Fatalf("StartWorkflow error: %v", err)
	}
	if inst.CurrentStateID != "start" {
		t.Errorf("expected current state 'start', got %q", inst.CurrentStateID)
	}
	if inst.Status != InstanceRunning {
		t.Errorf("expected RUNNING, got %q", inst.Status)
	}
}

func TestInitializeWorkflow(t *testing.T) {
	repos := newFakeRepos()
	def := padelDef()
	_ = repos.Workflows.Save(context.Background(), def)
	instanceRepo := repos.Instances.(*fakeInstanceRepo)
	instanceRepo.insts = map[string]*WorkflowInstance{
		"inst-1": {
			ID:                "inst-1",
			TenantID:          "tenant-1",
			ProjectID:         def.ProjectID,
			WorkflowID:        def.Slug,
			WorkflowVersionID: "version-1",
			Status:            InstanceCreated,
			Version:           0,
		},
	}
	eng := NewEngine(repos)

	inst, err := eng.InitializeWorkflow(context.Background(), "tenant-1", "inst-1")
	if err != nil {
		t.Fatalf("InitializeWorkflow error: %v", err)
	}
	if inst.CurrentStateID != def.EntryNodeID {
		t.Errorf("expected entry state %q, got %q", def.EntryNodeID, inst.CurrentStateID)
	}
	if inst.Status != InstanceRunning {
		t.Errorf("expected RUNNING, got %q", inst.Status)
	}
	if inst.Version != 1 {
		t.Errorf("expected version 1, got %d", inst.Version)
	}
}

func TestProcessEventHappyPath(t *testing.T) {
	repos := newFakeRepos()
	def := padelDef()
	_ = repos.Workflows.Save(context.Background(), def)
	eng := NewEngine(repos)
	inst, _ := eng.StartWorkflow(context.Background(), "t", "project-padel", "c", def, "workflow.started")

	// move start -> select_time
	_, _, err := eng.ProcessEvent(context.Background(), "t", inst.ID, &Event{ID: "e0", Type: "workflow.started", Source: SourceSystem})
	if err != nil {
		t.Fatalf("workflow.started error: %v", err)
	}

	// move to check_stock
	_, _, err = eng.ProcessEvent(context.Background(), "t", inst.ID, &Event{
		ID: "e1", Type: "datetime.selected", Source: SourceUser, Payload: map[string]any{"booking.date": "2026-08-27", "booking.time": "19:00"},
	})
	if err != nil {
		t.Fatalf("datetime.selected error: %v", err)
	}

	// slot available → confirm
	inst2, _, err := eng.ProcessEvent(context.Background(), "t", inst.ID, &Event{
		ID: "e2", Type: "slot.available", Source: SourceMCP, Payload: map[string]any{"slot.available": true},
	})
	if err != nil {
		t.Fatalf("slot.available error: %v", err)
	}
	if inst2.CurrentStateID != "confirm" {
		t.Errorf("expected confirm, got %q", inst2.CurrentStateID)
	}
}

func TestProcessEventTerminalMarksInstanceCompleted(t *testing.T) {
	repos := newFakeRepos()
	def := &WorkflowDefinition{
		Slug: "terminal", ProjectID: "project", EntryNodeID: "start",
		Nodes: []WorkflowNode{
			{ID: "start", Kind: NodeKindStart, Name: "START"},
			{ID: "done", Kind: NodeKindEnd, Name: "DONE", IsTerminal: true},
		},
		Transitions: []TransitionDefinition{{
			ID: "finish", SourceStateID: "start", Event: "finish", TargetStateID: "done", Priority: 1,
		}},
	}
	if err := repos.Workflows.Save(context.Background(), def); err != nil {
		t.Fatalf("save workflow: %v", err)
	}
	eng := NewEngine(repos)
	inst, err := eng.StartWorkflow(context.Background(), "tenant", "project", "conversation", def, "workflow.started")
	if err != nil {
		t.Fatalf("start workflow: %v", err)
	}

	completed, _, err := eng.ProcessEvent(context.Background(), "tenant", inst.ID, &Event{ID: "finish-1", Type: "finish", Source: SourceMCP})
	if err != nil {
		t.Fatalf("finish workflow: %v", err)
	}
	if completed.Status != InstanceCompleted {
		t.Fatalf("expected COMPLETED, got %s", completed.Status)
	}
}

func TestProcessEventGuardFail(t *testing.T) {
	repos := newFakeRepos()
	def := padelDef()
	_ = repos.Workflows.Save(context.Background(), def)
	eng := NewEngine(repos)
	inst, _ := eng.StartWorkflow(context.Background(), "t", "project-padel", "c", def, "workflow.started")
	_, _, _ = eng.ProcessEvent(context.Background(), "t", inst.ID, &Event{ID: "e0", Type: "workflow.started", Source: SourceSystem})
	_, _, _ = eng.ProcessEvent(context.Background(), "t", inst.ID, &Event{ID: "e1", Type: "datetime.selected", Source: SourceUser})

	// slot unavailable → guard t3 fails, t4 passes → back to start
	inst2, _, err := eng.ProcessEvent(context.Background(), "t", inst.ID, &Event{
		ID: "e2", Type: "slot.unavailable", Source: SourceMCP, Payload: map[string]any{"slot.available": false},
	})
	if err != nil {
		t.Fatalf("slot.unavailable error: %v", err)
	}
	if inst2.CurrentStateID != "start" {
		t.Errorf("expected back to start, got %q", inst2.CurrentStateID)
	}
}

func TestIdempotencyDedup(t *testing.T) {
	repos := newFakeRepos()
	def := padelDef()
	_ = repos.Workflows.Save(context.Background(), def)
	eng := NewEngine(repos)
	inst, _ := eng.StartWorkflow(context.Background(), "t", "project-padel", "c", def, "workflow.started")
	_, _, _ = eng.ProcessEvent(context.Background(), "t", inst.ID, &Event{ID: "e0", Type: "workflow.started", Source: SourceSystem})

	evt := &Event{ID: "e1", Type: "datetime.selected", Source: SourceUser, IdempotencyKey: "k1"}
	instA, _, err := eng.ProcessEvent(context.Background(), "t", inst.ID, evt)
	if err != nil {
		t.Fatalf("first event: %v", err)
	}
	// duplicate
	instB, _, err := eng.ProcessEvent(context.Background(), "t", inst.ID, evt)
	if err != nil {
		t.Fatalf("duplicate event: %v", err)
	}
	if instB.CurrentStateID != instA.CurrentStateID {
		t.Errorf("duplicate should be no-op, state changed unexpectedly")
	}
}

func TestIntentResolver(t *testing.T) {
	reg := IntentRegistry{SchemaVersion: 1, Intents: []IntentDefinition{
		{ID: "BOOKING_PADEL", ProjectID: "project-padel", Name: "Padel", WorkflowSlug: "padel-booking", EntryEvent: "padel.booking.requested", Priority: 10},
	}}
	lookup := func(projectID, slug string) (*WorkflowDefinition, bool) {
		if projectID == "project-padel" && slug == "padel-booking" {
			return padelDef(), true
		}
		return nil, false
	}
	r := NewIntentResolver(reg, lookup)
	def, entry, stateID, ok := r.ResolveIntent("project-padel", "BOOKING_PADEL")
	if !ok {
		t.Fatal("intent not resolved")
	}
	if def.Slug != "padel-booking" || entry != "padel.booking.requested" || stateID != "start" {
		t.Errorf("unexpected resolution: %s %s %s", def.Slug, entry, stateID)
	}
}
