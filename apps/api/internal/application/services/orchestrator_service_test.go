package services

import (
	"context"
	"database/sql"
	"testing"

	"github.com/irosadie/open-state/api/internal/domain/engine"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	domain "github.com/irosadie/open-state/go-shared/domain"
)

// fakeInstanceRepo is a minimal in-memory IInstanceRepository for orchestrator tests.
type fakeInstanceRepo struct {
	instances map[string]*entities.WorkflowInstance
}

func (f *fakeInstanceRepo) Create(_ context.Context, _ string, input repositories.CreateWorkflowInstanceInput) (*entities.WorkflowInstance, error) {
	inst := &entities.WorkflowInstance{
		ID:            "inst-1",
		WorkflowID:    input.WorkflowID,
		WorkflowVersionID: input.WorkflowVersionID,
		Status:        entities.WorkflowInstanceRunning,
		Version:       1,
	}
	if input.CorrelationKey != nil {
		inst.CorrelationKey = sql.NullString{String: *input.CorrelationKey, Valid: true}
	}
	if f.instances == nil {
		f.instances = map[string]*entities.WorkflowInstance{}
	}
	f.instances[inst.ID] = inst
	return inst, nil
}

func (f *fakeInstanceRepo) FindByID(_ context.Context, _ string, id string) (*entities.WorkflowInstance, error) {
	inst, ok := f.instances[id]
	if !ok {
		return nil, domain.NewNotFound("workflow instance not found")
	}
	return inst, nil
}

func (f *fakeInstanceRepo) UpdateStatus(_ context.Context, _ string, id string, status entities.WorkflowInstanceStatus, expectedVersion int) (*entities.WorkflowInstance, error) {
	inst, ok := f.instances[id]
	if !ok {
		return nil, domain.NewNotFound("workflow instance not found")
	}
	if inst.Version != expectedVersion {
		return nil, domain.NewConflict("optimistic lock conflict")
	}
	inst.Status = status
	inst.Version++
	return inst, nil
}

func (f *fakeInstanceRepo) ListByTenant(_ context.Context, _ string) ([]entities.WorkflowInstance, error) {
	out := make([]entities.WorkflowInstance, 0, len(f.instances))
	for _, i := range f.instances {
		out = append(out, *i)
	}
	return out, nil
}

func (f *fakeInstanceRepo) ListStateInstancesByWorkflowInstance(_ context.Context, _ string, _ string) ([]entities.StateInstance, error) {
	return nil, nil
}

func (f *fakeInstanceRepo) Transition(context.Context, string, repositories.TransitionInput) (*entities.StateInstance, error) {
	return nil, nil
}
func (f *fakeInstanceRepo) CreateStateInstance(context.Context, string, repositories.CreateStateInstanceInput) (*entities.StateInstance, error) {
	return nil, nil
}
func (f *fakeInstanceRepo) FindStateInstanceByID(context.Context, string, string) (*entities.StateInstance, error) {
	return nil, nil
}
func (f *fakeInstanceRepo) UpdateStateInstanceStatus(context.Context, string, string, entities.StateInstanceStatus, int) (*entities.StateInstance, error) {
	return nil, nil
}
func (f *fakeInstanceRepo) IncrementRetry(context.Context, string, string, int) (*entities.StateInstance, error) {
	return nil, nil
}

// fakeEventRepo is a minimal in-memory IEventRepository for orchestrator tests.
type fakeEventRepo struct {
	events []entities.Event
}

func (f *fakeEventRepo) Append(_ context.Context, _ string, input repositories.AppendEventInput) (*entities.Event, error) {
	evt := &entities.Event{
		ID:      "evt-1",
		EventID: input.EventID,
		Type:    input.Type,
		Source:  input.Source,
		WorkflowInstanceID: input.WorkflowInstanceID,
		Timestamp: input.Timestamp,
		Payload:  input.Payload,
		Sequence: int64(len(f.events) + 1),
	}
	f.events = append(f.events, *evt)
	return evt, nil
}

func (f *fakeEventRepo) ListEventsByInstance(_ context.Context, _ string, _ string) ([]entities.Event, error) {
	return f.events, nil
}

func (f *fakeEventRepo) FindEventByID(context.Context, string, string) (*entities.Event, error) { return nil, nil }
func (f *fakeEventRepo) ListEventsByTenant(context.Context, string) ([]entities.Event, error)   { return nil, nil }
func (f *fakeEventRepo) InsertInbox(context.Context, string, repositories.InsertInboxEventInput) (*entities.InboxEvent, error) {
	return nil, nil
}
func (f *fakeEventRepo) ClaimInbox(context.Context, string, int) ([]entities.InboxEvent, error) { return nil, nil }
func (f *fakeEventRepo) MarkInboxProcessed(context.Context, string, string) (*entities.InboxEvent, error) {
	return nil, nil
}
func (f *fakeEventRepo) InsertOutbox(context.Context, string, repositories.InsertOutboxEventInput) (*entities.OutboxEvent, error) {
	return nil, nil
}
func (f *fakeEventRepo) ClaimOutbox(context.Context, string, int) ([]entities.OutboxEvent, error) { return nil, nil }
func (f *fakeEventRepo) MarkOutboxPublished(context.Context, string, string) (*entities.OutboxEvent, error) {
	return nil, nil
}
func (f *fakeEventRepo) UpsertIdempotency(context.Context, string, repositories.UpsertIdempotencyInput) (*entities.IdempotencyRecord, error) {
	return nil, nil
}
func (f *fakeEventRepo) FindIdempotency(context.Context, string, string) (*entities.IdempotencyRecord, error) {
	return nil, nil
}

func newOrchestratorForTest() (*OrchestratorService, *fakeInstanceRepo, *fakeEventRepo) {
	instRepo := &fakeInstanceRepo{}
	evtRepo := &fakeEventRepo{}
	svc := NewOrchestratorService(instRepo, evtRepo, &fakeContextRepo{}, &fakeCapRepo{})
	return svc, instRepo, evtRepo
}

func TestOrchestratorStartAndSuspendResume(t *testing.T) {
	svc, _, _ := newOrchestratorForTest()
	ctx := context.Background()

	created, err := svc.StartWorkflow(ctx, "tenant-1", "wf-1", "wv-1", "conv-1")
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	if created.Status != entities.WorkflowInstanceRunning {
		t.Fatalf("expected RUNNING, got %s", created.Status)
	}

	suspended, err := svc.SuspendWorkflow(ctx, "tenant-1", created.ID)
	if err != nil {
		t.Fatalf("suspend: %v", err)
	}
	if suspended.Status != entities.WorkflowInstanceSuspended {
		t.Fatalf("expected SUSPENDED, got %s", suspended.Status)
	}

	resumed, err := svc.ResumeWorkflow(ctx, "tenant-1", created.ID)
	if err != nil {
		t.Fatalf("resume: %v", err)
	}
	if resumed.Status != entities.WorkflowInstanceRunning {
		t.Fatalf("expected RUNNING after resume, got %s", resumed.Status)
	}
}

func TestOrchestratorCancelTerminalConflict(t *testing.T) {
	svc, instRepo, _ := newOrchestratorForTest()
	ctx := context.Background()

	created, _ := svc.StartWorkflow(ctx, "tenant-1", "wf-1", "wv-1", "")
	instRepo.instances[created.ID].Status = entities.WorkflowInstanceCompleted

	if _, err := svc.CancelWorkflow(ctx, "tenant-1", created.ID); err == nil {
		t.Fatal("expected conflict for terminal instance")
	}
}

func TestOrchestratorProposeEvent(t *testing.T) {
	svc, _, evtRepo := newOrchestratorForTest()
	ctx := context.Background()

	created, _ := svc.StartWorkflow(ctx, "tenant-1", "wf-1", "wv-1", "conv-1")
	evt, err := svc.ProposeEvent(ctx, "tenant-1", created.ID, "payment.success", map[string]any{"amount": 100}, "corr-1")
	if err != nil {
		t.Fatalf("propose: %v", err)
	}
	if evt.Type != "payment.success" {
		t.Fatalf("expected event type, got %s", evt.Type)
	}
	if evt.Source != entities.EventSourceMCP {
		t.Fatalf("expected MCP source, got %s", evt.Source)
	}
	if len(evtRepo.events) != 1 {
		t.Fatalf("expected 1 event appended, got %d", len(evtRepo.events))
	}

	// Empty event type must be rejected.
	if _, err := svc.ProposeEvent(ctx, "tenant-1", created.ID, "", nil, ""); err == nil {
		t.Fatal("expected validation error for empty event type")
	}
}

func TestOrchestratorListHistory(t *testing.T) {
	svc, _, _ := newOrchestratorForTest()
	ctx := context.Background()

	created, _ := svc.StartWorkflow(ctx, "tenant-1", "wf-1", "wv-1", "conv-1")
	_, _ = svc.ProposeEvent(ctx, "tenant-1", created.ID, "state.entered", nil, "")

	history, err := svc.ListHistory(ctx, "tenant-1", created.ID)
	if err != nil {
		t.Fatalf("history: %v", err)
	}
	if len(history) != 1 {
		t.Fatalf("expected 1 history event, got %d", len(history))
	}
}

func TestOrchestratorGetActiveWorkflow(t *testing.T) {
	svc, _, _ := newOrchestratorForTest()
	ctx := context.Background()

	created, _ := svc.StartWorkflow(ctx, "tenant-1", "wf-1", "wv-1", "conv-1")

	active, err := svc.GetActiveWorkflow(ctx, "tenant-1", "conv-1")
	if err != nil {
		t.Fatalf("get active: %v", err)
	}
	if active.ID != created.ID {
		t.Fatalf("expected active instance %s, got %s", created.ID, active.ID)
	}

	// Terminal instances must not be considered active.
	_, _ = svc.CancelWorkflow(ctx, "tenant-1", created.ID)
	if _, err := svc.GetActiveWorkflow(ctx, "tenant-1", "conv-1"); err == nil {
		t.Fatal("expected not-found after cancel")
	}
}

func TestOrchestratorReplayWorkflow(t *testing.T) {
	svc, _, _ := newOrchestratorForTest()
	ctx := context.Background()

	created, _ := svc.StartWorkflow(ctx, "tenant-1", "wf-1", "wv-1", "conv-1")
	_, _ = svc.ProposeEvent(ctx, "tenant-1", created.ID, "slot.booked", map[string]any{"court": "A"}, "")
	_, _ = svc.ProposeEvent(ctx, "tenant-1", created.ID, "slot.paid", map[string]any{"amount": 100}, "")

	snapshot, last, err := svc.ReplayWorkflow(ctx, "tenant-1", created.ID)
	if err != nil {
		t.Fatalf("replay: %v", err)
	}
	if snapshot["court"] != "A" {
		t.Fatalf("expected replayed court=A, got %v", snapshot["court"])
	}
	if snapshot["amount"] != float64(100) {
		t.Fatalf("expected replayed amount=100, got %v", snapshot["amount"])
	}
	if last == nil || last.Type != "slot.paid" {
		t.Fatalf("expected last event slot.paid, got %+v", last)
	}
}

// ---- Engine-backed propose test ------------------------------------------

// fakeEngineWorkflowRepo satisfies engine.WorkflowRepository with an in-memory def.
type fakeEngineWorkflowRepo struct{ defs map[string]*engine.WorkflowDefinition }

func (f *fakeEngineWorkflowRepo) GetBySlug(_ context.Context, _, projectID, slug string) (*engine.WorkflowDefinition, error) {
	d, ok := f.defs[projectID+"/"+slug]
	if !ok {
		return nil, domain.NewNotFound("workflow not found")
	}
	return d, nil
}
func (f *fakeEngineWorkflowRepo) Save(_ context.Context, d *engine.WorkflowDefinition) error {
	if f.defs == nil {
		f.defs = map[string]*engine.WorkflowDefinition{}
	}
	f.defs[d.ProjectID+"/"+d.Slug] = d
	return nil
}

// fakeEngineInstanceRepo tracks the engine instance's current state.
type fakeEngineInstanceRepo struct {
	instances map[string]*engine.WorkflowInstance
}

func (f *fakeEngineInstanceRepo) Create(_ context.Context, i *engine.WorkflowInstance) error {
	if f.instances == nil {
		f.instances = map[string]*engine.WorkflowInstance{}
	}
	cp := *i
	f.instances[i.ID] = &cp
	return nil
}
func (f *fakeEngineInstanceRepo) Get(_ context.Context, _, id string) (*engine.WorkflowInstance, error) {
	i, ok := f.instances[id]
	if !ok {
		return nil, domain.NewNotFound("instance not found")
	}
	cp := *i
	return &cp, nil
}
func (f *fakeEngineInstanceRepo) UpdateWithVersion(_ context.Context, i *engine.WorkflowInstance, expected int) error {
	cur, ok := f.instances[i.ID]
	if !ok {
		return domain.NewNotFound("instance not found")
	}
	if cur.Version != expected {
		return domain.NewConflict("version conflict")
	}
	cp := *i
	f.instances[i.ID] = &cp
	return nil
}

// fakeEngineEventRepo records engine events + idempotency.
type fakeEngineEventRepo struct {
	processed map[string]bool
}

func (f *fakeEngineEventRepo) Append(context.Context, *engine.Event) error { return nil }
func (f *fakeEngineEventRepo) IsProcessed(_ context.Context, _, key string) (bool, error) {
	return f.processed[key], nil
}
func (f *fakeEngineEventRepo) MarkProcessed(_ context.Context, _, key, _ string) error {
	if f.processed == nil {
		f.processed = map[string]bool{}
	}
	f.processed[key] = true
	return nil
}

// fakeEngineProjectRepo satisfies engine.ProjectRepository.
type fakeEngineProjectRepo struct{}

func (fakeEngineProjectRepo) Get(context.Context, string, string) (*engine.Project, error) {
	return &engine.Project{ID: "proj"}, nil
}
func (fakeEngineProjectRepo) Save(context.Context, *engine.Project) error { return nil }

func engineBackedOrchestrator() (*OrchestratorService, *fakeEngineInstanceRepo) {
	def := &engine.WorkflowDefinition{
		Slug: "wf", ProjectID: "proj", Name: "WF", SchemaVersion: 1,
		Status: engine.WorkflowPublished, EntryNodeID: "start",
		Nodes: []engine.WorkflowNode{
			{ID: "start", Kind: engine.NodeKindStart, Name: "START"},
			{ID: "s1", Kind: engine.NodeKindState, Name: "S1"},
		},
		Transitions: []engine.TransitionDefinition{
			{ID: "t0", SourceStateID: "start", Event: "go", TargetStateID: "s1", Priority: 1},
		},
		Policy: engine.WorkflowPolicy{Interruptible: "NEVER", Priority: 1},
	}
	wfRepo := &fakeEngineWorkflowRepo{}
	_ = wfRepo.Save(context.Background(), def)
	instRepo := &fakeEngineInstanceRepo{}
	eng := engine.NewEngine(engine.EngineRepositories{
		Projects:  fakeEngineProjectRepo{},
		Workflows: wfRepo,
		Instances: instRepo,
		Events:    &fakeEngineEventRepo{},
	})
	// The engine-backed orchestrator uses persistence fakes for the pre-checks,
	// and the in-memory engine repos for the actual transition.
	persistInstRepo := &fakeInstanceRepo{}
	evtRepo := &fakeEventRepo{}
	svc := NewEngineOrchestratorService(persistInstRepo, evtRepo, &fakeContextRepo{}, &fakeCapRepo{}, eng)
	return svc, instRepo
}

func TestOrchestratorProposeEventEngineBacked(t *testing.T) {
	svc, engineInstRepo := engineBackedOrchestrator()
	ctx := context.Background()

	// Create a persisted instance (active) so ProposeEvent's pre-check passes.
	created, err := svc.StartWorkflow(ctx, "tenant-1", "wf-1", "wv-1", "conv-1")
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	// Register the same instance id in the engine instance repo at "start".
	_ = engineInstRepo.Create(ctx, &engine.WorkflowInstance{
		ID: created.ID, TenantID: "tenant-1", WorkflowID: "wf", ProjectID: "proj",
		WorkflowVersionID: "wv-1", Status: engine.InstanceRunning, CurrentStateID: "start", Version: 1,
	})

	// A valid event "go" transitions start -> s1.
	evt, err := svc.ProposeEvent(ctx, "tenant-1", created.ID, "go", nil, "")
	if err != nil {
		t.Fatalf("engine-backed propose: %v", err)
	}
	if evt.Type != "go" {
		t.Fatalf("expected event type go, got %s", evt.Type)
	}
	got, _ := engineInstRepo.Get(ctx, "tenant-1", created.ID)
	if got.CurrentStateID != "s1" {
		t.Fatalf("expected engine state s1, got %q", got.CurrentStateID)
	}
}

func TestOrchestratorProposeEventEngineBackedRejects(t *testing.T) {
	svc, engineInstRepo := engineBackedOrchestrator()
	ctx := context.Background()

	created, _ := svc.StartWorkflow(ctx, "tenant-1", "wf-1", "wv-1", "conv-1")
	_ = engineInstRepo.Create(ctx, &engine.WorkflowInstance{
		ID: created.ID, TenantID: "tenant-1", WorkflowID: "wf", ProjectID: "proj",
		WorkflowVersionID: "wv-1", Status: engine.InstanceRunning, CurrentStateID: "start", Version: 1,
	})

	// "not_allowed" is not a valid event from "start" -> engine must reject.
	if _, err := svc.ProposeEvent(ctx, "tenant-1", created.ID, "not_allowed", nil, ""); err == nil {
		t.Fatal("expected engine rejection for disallowed event")
	}
	got, _ := engineInstRepo.Get(ctx, "tenant-1", created.ID)
	if got.CurrentStateID != "start" {
		t.Fatalf("expected state unchanged at start, got %q", got.CurrentStateID)
	}
}

var _ repositories.IInstanceRepository = (*fakeInstanceRepo)(nil)
var _ repositories.IEventRepository = (*fakeEventRepo)(nil)
