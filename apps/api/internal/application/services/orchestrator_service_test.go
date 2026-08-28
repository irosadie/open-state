package services

import (
	"context"
	"testing"

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

var _ repositories.IInstanceRepository = (*fakeInstanceRepo)(nil)
var _ repositories.IEventRepository = (*fakeEventRepo)(nil)
