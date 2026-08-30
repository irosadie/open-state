package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	domain "github.com/irosadie/open-state/go-shared/domain"
)

type inspectorInstanceRepo struct {
	tenant   string
	instance *entities.WorkflowInstance
}

func (f *inspectorInstanceRepo) Create(context.Context, string, repositories.CreateWorkflowInstanceInput) (*entities.WorkflowInstance, error) {
	return nil, nil
}
func (f *inspectorInstanceRepo) FindByID(_ context.Context, tenantID, id string) (*entities.WorkflowInstance, error) {
	if tenantID != f.tenant || f.instance == nil || f.instance.ID != id {
		return nil, domain.NewNotFound("workflow instance not found")
	}
	return f.instance, nil
}
func (f *inspectorInstanceRepo) ListByTenant(_ context.Context, tenantID string) ([]entities.WorkflowInstance, error) {
	if tenantID != f.tenant || f.instance == nil {
		return []entities.WorkflowInstance{}, nil
	}
	return []entities.WorkflowInstance{*f.instance}, nil
}
func (f *inspectorInstanceRepo) UpdateStatus(context.Context, string, string, entities.WorkflowInstanceStatus, int) (*entities.WorkflowInstance, error) {
	return nil, nil
}
func (f *inspectorInstanceRepo) Transition(context.Context, string, repositories.TransitionInput) (*entities.StateInstance, error) {
	return nil, nil
}
func (f *inspectorInstanceRepo) CreateStateInstance(context.Context, string, repositories.CreateStateInstanceInput) (*entities.StateInstance, error) {
	return nil, nil
}
func (f *inspectorInstanceRepo) FindStateInstanceByID(context.Context, string, string) (*entities.StateInstance, error) {
	return nil, nil
}
func (f *inspectorInstanceRepo) UpdateStateInstanceStatus(context.Context, string, string, entities.StateInstanceStatus, int) (*entities.StateInstance, error) {
	return nil, nil
}
func (f *inspectorInstanceRepo) IncrementRetry(context.Context, string, string, int) (*entities.StateInstance, error) {
	return nil, nil
}

type inspectorReadRepo struct {
	workflow       *entities.Workflow
	version        *entities.WorkflowVersion
	states         []entities.State
	stateInstances []entities.StateInstance
}

func (f *inspectorReadRepo) FindWorkflow(context.Context, string, string) (*entities.Workflow, error) {
	return f.workflow, nil
}
func (f *inspectorReadRepo) FindWorkflowVersion(context.Context, string, string) (*entities.WorkflowVersion, error) {
	return f.version, nil
}
func (f *inspectorReadRepo) ListStatesByVersion(context.Context, string, string) ([]entities.State, error) {
	return f.states, nil
}
func (f *inspectorReadRepo) ListStateInstancesByWorkflowInstance(context.Context, string, string) ([]entities.StateInstance, error) {
	return f.stateInstances, nil
}

type inspectorEventRepo struct{ events []entities.Event }

func (f *inspectorEventRepo) Append(context.Context, string, repositories.AppendEventInput) (*entities.Event, error) {
	return nil, nil
}
func (f *inspectorEventRepo) FindEventByID(context.Context, string, string) (*entities.Event, error) {
	return nil, nil
}
func (f *inspectorEventRepo) ListEventsByInstance(context.Context, string, string) ([]entities.Event, error) {
	return f.events, nil
}
func (f *inspectorEventRepo) ListEventsByTenant(context.Context, string) ([]entities.Event, error) {
	return f.events, nil
}
func (f *inspectorEventRepo) InsertInbox(context.Context, string, repositories.InsertInboxEventInput) (*entities.InboxEvent, error) {
	return nil, nil
}
func (f *inspectorEventRepo) ClaimInbox(context.Context, string, int) ([]entities.InboxEvent, error) {
	return nil, nil
}
func (f *inspectorEventRepo) MarkInboxProcessed(context.Context, string, string) (*entities.InboxEvent, error) {
	return nil, nil
}
func (f *inspectorEventRepo) InsertOutbox(context.Context, string, repositories.InsertOutboxEventInput) (*entities.OutboxEvent, error) {
	return nil, nil
}
func (f *inspectorEventRepo) ClaimOutbox(context.Context, string, int) ([]entities.OutboxEvent, error) {
	return nil, nil
}
func (f *inspectorEventRepo) MarkOutboxPublished(context.Context, string, string) (*entities.OutboxEvent, error) {
	return nil, nil
}
func (f *inspectorEventRepo) UpsertIdempotency(context.Context, string, repositories.UpsertIdempotencyInput) (*entities.IdempotencyRecord, error) {
	return nil, nil
}
func (f *inspectorEventRepo) FindIdempotency(context.Context, string, string) (*entities.IdempotencyRecord, error) {
	return nil, nil
}

type inspectorAuditRepo struct{ entries []entities.AuditLog }

func (f *inspectorAuditRepo) Append(context.Context, string, repositories.AppendAuditLogInput) (*entities.AuditLog, error) {
	return nil, nil
}
func (f *inspectorAuditRepo) ListByTenant(context.Context, string) ([]entities.AuditLog, error) {
	return f.entries, nil
}
func (f *inspectorAuditRepo) ListByAction(context.Context, string, entities.AuditAction) ([]entities.AuditLog, error) {
	return f.entries, nil
}
func (f *inspectorAuditRepo) ListByResource(context.Context, string, string, string) ([]entities.AuditLog, error) {
	return f.entries, nil
}
func (f *inspectorAuditRepo) ListFiltered(context.Context, string, repositories.AuditFilter) ([]entities.AuditLog, error) {
	return f.entries, nil
}
func (f *inspectorAuditRepo) CountFiltered(context.Context, string, repositories.AuditFilter) (int64, error) {
	return int64(len(f.entries)), nil
}

func TestRuntimeInspectorComposesOrderedSanitizedDetail(t *testing.T) {
	instance := &entities.WorkflowInstance{
		ID: "instance-1", TenantID: "tenant-a", WorkflowID: "workflow-1", WorkflowVersionID: "version-1",
		CorrelationKey: sql.NullString{String: "conversation-1", Valid: true}, Status: entities.WorkflowInstanceRunning,
		CurrentStateInstanceID: stringPtr("state-instance-2"), CreatedAt: time.Date(2026, 8, 29, 10, 0, 0, 0, time.UTC), UpdatedAt: time.Date(2026, 8, 29, 10, 2, 0, 0, time.UTC),
	}
	read := &inspectorReadRepo{
		workflow: &entities.Workflow{ID: "workflow-1", Name: "Booking", Slug: "booking"},
		version:  &entities.WorkflowVersion{ID: "version-1", VersionNo: 3},
		states:   []entities.State{{Key: "START", Name: "Start", RequiredContext: json.RawMessage(`["customer.email","slot"]`)}, {Key: "CONFIRM", Name: "Confirm", RequiredContext: json.RawMessage(`["slot","payment"]`)}},
		stateInstances: []entities.StateInstance{
			{ID: "state-instance-1", StateKey: "START", Status: entities.StateInstanceCompleted, EnteredAt: time.Date(2026, 8, 29, 10, 0, 1, 0, time.UTC)},
			{ID: "state-instance-2", StateKey: "CONFIRM", Status: entities.StateInstanceActive, EnteredAt: time.Date(2026, 8, 29, 10, 1, 0, 0, time.UTC)},
		},
	}
	events := &inspectorEventRepo{events: []entities.Event{
		{ID: "event-2", Type: "payment.received", Sequence: 2, Timestamp: time.Date(2026, 8, 29, 10, 1, 30, 0, time.UTC)},
		{ID: "event-1", Type: "booking.started", Sequence: 1, Timestamp: time.Date(2026, 8, 29, 10, 0, 30, 0, time.UTC)},
	}}
	contexts := &fakeContextRepo{scopes: map[string][]entities.ContextRecord{
		"conversation-1": {{Key: "customer.email", Value: json.RawMessage(`"user@example.com"`)}},
		"instance-1":     {{Key: "slot", Value: json.RawMessage(`"court-a"`)}},
	}}
	audits := &inspectorAuditRepo{entries: []entities.AuditLog{{ID: "audit-1", Action: entities.AuditActionGuardFailed, ResourceType: "workflow_instance", ResourceID: "instance-1", OccurredAt: time.Date(2026, 8, 29, 10, 1, 45, 0, time.UTC)}}}
	svc := NewRuntimeInspectorService(&inspectorInstanceRepo{tenant: "tenant-a", instance: instance}, read, events, contexts, audits, nil)

	got, err := svc.Get(context.Background(), "tenant-a", "instance-1")
	if err != nil {
		t.Fatalf("get detail: %v", err)
	}
	if got.Summary.Workflow.Name != "Booking" || got.Summary.Workflow.Version != 3 {
		t.Fatalf("workflow summary not composed: %#v", got.Summary.Workflow)
	}
	if got.Context.Available["customer.email"] != "[REDACTED]" {
		t.Fatalf("sensitive context was retained: %#v", got.Context.Available)
	}
	if len(got.Context.Missing) != 1 || got.Context.Missing[0] != "payment" {
		t.Fatalf("unexpected missing context: %#v", got.Context.Missing)
	}
	if len(got.Timeline) < 3 || got.Timeline[0].ID != "state-instance-1" || got.Timeline[1].ID != "event-1" {
		t.Fatalf("timeline was not ordered chronologically: %#v", got.Timeline)
	}
	if got.Timeline[len(got.Timeline)-1].ReasonCode == nil || *got.Timeline[len(got.Timeline)-1].ReasonCode != "GUARD_FAILED" {
		t.Fatalf("guard evidence missing: %#v", got.Timeline)
	}
}

func TestRuntimeInspectorDoesNotExposeOtherTenant(t *testing.T) {
	svc := NewRuntimeInspectorService(&inspectorInstanceRepo{tenant: "tenant-a", instance: &entities.WorkflowInstance{ID: "instance-1"}}, nil, nil, nil, nil, nil)
	if _, err := svc.Get(context.Background(), "tenant-b", "instance-1"); err == nil {
		t.Fatal("expected cross-tenant lookup to fail")
	}
}

func TestRuntimeInspectorReturnsPartialTraceAsAvailableEvidence(t *testing.T) {
	traceRepo := &recorderTraceRepo{entries: []entities.RuntimeTraceEntry{{
		ID:                 "trace-1",
		WorkflowInstanceID: "instance-1",
		Stage:              entities.RuntimeTraceStageEventHandling,
		Source:             entities.RuntimeTraceSourceOpenState,
		Status:             entities.RuntimeTraceStatusSucceeded,
	}}}
	svc := NewRuntimeInspectorService(
		&inspectorInstanceRepo{tenant: "tenant-a", instance: &entities.WorkflowInstance{ID: "instance-1"}},
		nil, nil, nil, nil, traceRepo,
	)

	got, err := svc.DebugTrace(context.Background(), "tenant-a", "instance-1", "")
	if err != nil {
		t.Fatalf("get partial trace: %v", err)
	}
	if !got.Available || len(got.Data) != 1 || got.Data[0].Stage != string(entities.RuntimeTraceStageEventHandling) {
		t.Fatalf("unexpected partial trace response: %+v", got)
	}
}

func stringPtr(value string) *string { return &value }
