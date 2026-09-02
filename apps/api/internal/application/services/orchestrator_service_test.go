package services

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/irosadie/open-state/api/internal/domain/engine"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	domain "github.com/irosadie/open-state/go-shared/domain"
)

// fakeInstanceRepo is a minimal in-memory IInstanceRepository for orchestrator tests.
type fakeInstanceRepo struct {
	instances    map[string]*entities.WorkflowInstance
	createStatus entities.WorkflowInstanceStatus
	createCalls  int
}

func (f *fakeInstanceRepo) Create(_ context.Context, tenantID string, input repositories.CreateWorkflowInstanceInput) (*entities.WorkflowInstance, error) {
	f.createCalls++
	status := f.createStatus
	if status == "" {
		status = entities.WorkflowInstanceRunning
	}
	version := 1
	if status == entities.WorkflowInstanceCreated {
		version = 0
	}
	inst := &entities.WorkflowInstance{
		ID:                "inst-1",
		TenantID:          tenantID,
		WorkflowID:        input.WorkflowID,
		WorkflowVersionID: input.WorkflowVersionID,
		Status:            status,
		Version:           version,
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

func (f *fakeInstanceRepo) FindByID(_ context.Context, tenantID, id string) (*entities.WorkflowInstance, error) {
	inst, ok := f.instances[id]
	if !ok || (inst.TenantID != "" && inst.TenantID != tenantID) {
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
	events      []entities.Event
	idempotency map[string]*entities.IdempotencyRecord
}

func (f *fakeEventRepo) Append(_ context.Context, _ string, input repositories.AppendEventInput) (*entities.Event, error) {
	var idempotencyKey sql.NullString
	if input.IdempotencyKey != nil {
		idempotencyKey = sql.NullString{String: *input.IdempotencyKey, Valid: true}
	}
	evt := &entities.Event{
		ID:                 "evt-1",
		EventID:            input.EventID,
		Type:               input.Type,
		Source:             input.Source,
		WorkflowInstanceID: input.WorkflowInstanceID,
		Timestamp:          input.Timestamp,
		Payload:            input.Payload,
		Sequence:           int64(len(f.events) + 1),
		IdempotencyKey:     idempotencyKey,
	}
	f.events = append(f.events, *evt)
	return evt, nil
}

func (f *fakeEventRepo) ListEventsByInstance(_ context.Context, _ string, _ string) ([]entities.Event, error) {
	return f.events, nil
}

func (f *fakeEventRepo) FindEventByID(_ context.Context, _ string, id string) (*entities.Event, error) {
	for i := range f.events {
		if f.events[i].ID == id || f.events[i].EventID == id {
			return &f.events[i], nil
		}
	}
	return nil, domain.NewNotFound("event not found")
}
func (f *fakeEventRepo) ListEventsByTenant(context.Context, string) ([]entities.Event, error) {
	return nil, nil
}
func (f *fakeEventRepo) InsertInbox(context.Context, string, repositories.InsertInboxEventInput) (*entities.InboxEvent, error) {
	return nil, nil
}
func (f *fakeEventRepo) ClaimInbox(context.Context, string, int) ([]entities.InboxEvent, error) {
	return nil, nil
}
func (f *fakeEventRepo) MarkInboxProcessed(context.Context, string, string) (*entities.InboxEvent, error) {
	return nil, nil
}
func (f *fakeEventRepo) InsertOutbox(context.Context, string, repositories.InsertOutboxEventInput) (*entities.OutboxEvent, error) {
	return nil, nil
}
func (f *fakeEventRepo) ClaimOutbox(context.Context, string, int) ([]entities.OutboxEvent, error) {
	return nil, nil
}
func (f *fakeEventRepo) MarkOutboxPublished(context.Context, string, string) (*entities.OutboxEvent, error) {
	return nil, nil
}
func (f *fakeEventRepo) UpsertIdempotency(_ context.Context, tenantID string, input repositories.UpsertIdempotencyInput) (*entities.IdempotencyRecord, error) {
	if f.idempotency == nil {
		f.idempotency = map[string]*entities.IdempotencyRecord{}
	}
	record := &entities.IdempotencyRecord{
		ID:             input.IdempotencyKey,
		TenantID:       tenantID,
		IdempotencyKey: input.IdempotencyKey,
		Scope:          input.Scope,
		ResultStatus:   input.ResultStatus,
		Payload:        input.Payload,
	}
	f.idempotency[input.IdempotencyKey] = record
	return record, nil
}
func (f *fakeEventRepo) FindIdempotency(_ context.Context, tenantID, key string) (*entities.IdempotencyRecord, error) {
	record, ok := f.idempotency[key]
	if !ok || record.TenantID != tenantID {
		return nil, domain.NewNotFound("idempotency record not found")
	}
	return record, nil
}

// fakeCapabilityEvidenceRepo models the tenant/instance/state/idempotency
// dimensions used by the State MCP transition gate.
type fakeCapabilityEvidenceRepo struct {
	records []entities.CapabilityExecutionEvidence
}

func (f *fakeCapabilityEvidenceRepo) Upsert(_ context.Context, input repositories.CapabilityEvidenceInput) (*entities.CapabilityExecutionEvidence, error) {
	for i := range f.records {
		current := &f.records[i]
		if current.TenantID == input.TenantID && current.ProjectID == input.ProjectID &&
			current.WorkflowInstanceID == input.WorkflowInstanceID && current.StateID == input.StateID &&
			current.CapabilityID == input.CapabilityID && current.IdempotencyKey == input.IdempotencyKey {
			current.Status = input.Status
			current.Result = input.Result
			current.Error = input.Error
			return current, nil
		}
	}
	correlation := input.CorrelationID
	record := entities.CapabilityExecutionEvidence{
		ID:                 "evidence",
		TenantID:           input.TenantID,
		ProjectID:          input.ProjectID,
		WorkflowInstanceID: input.WorkflowInstanceID,
		StateID:            input.StateID,
		CapabilityID:       input.CapabilityID,
		CapabilityName:     input.CapabilityName,
		ProviderServer:     input.ProviderServer,
		ProviderTool:       input.ProviderTool,
		CorrelationID:      correlation,
		IdempotencyKey:     input.IdempotencyKey,
		Status:             input.Status,
		Result:             input.Result,
		Error:              input.Error,
	}
	f.records = append(f.records, record)
	return &f.records[len(f.records)-1], nil
}

func (f *fakeCapabilityEvidenceRepo) FindByIdempotency(_ context.Context, tenantID, projectID, instanceID, stateID, capabilityID, key string) (*entities.CapabilityExecutionEvidence, error) {
	for i := range f.records {
		record := &f.records[i]
		if record.TenantID == tenantID && record.ProjectID == projectID && record.WorkflowInstanceID == instanceID &&
			record.StateID == stateID && record.CapabilityID == capabilityID && record.IdempotencyKey == key {
			return record, nil
		}
	}
	return nil, domain.NewNotFound("capability evidence not found")
}

func (f *fakeCapabilityEvidenceRepo) ListByState(_ context.Context, tenantID, projectID, instanceID, stateID string) ([]entities.CapabilityExecutionEvidence, error) {
	return f.list(tenantID, projectID, instanceID, stateID, true), nil
}

func (f *fakeCapabilityEvidenceRepo) ListByInstanceState(_ context.Context, tenantID, instanceID, stateID string) ([]entities.CapabilityExecutionEvidence, error) {
	return f.list(tenantID, "", instanceID, stateID, false), nil
}

func (f *fakeCapabilityEvidenceRepo) list(tenantID, projectID, instanceID, stateID string, scopedProject bool) []entities.CapabilityExecutionEvidence {
	out := make([]entities.CapabilityExecutionEvidence, 0)
	for _, record := range f.records {
		if record.TenantID != tenantID || record.WorkflowInstanceID != instanceID || record.StateID != stateID {
			continue
		}
		if scopedProject && record.ProjectID != projectID {
			continue
		}
		out = append(out, record)
	}
	return out
}

func newOrchestratorForTest() (*OrchestratorService, *fakeInstanceRepo, *fakeEventRepo) {
	instRepo := &fakeInstanceRepo{}
	evtRepo := &fakeEventRepo{}
	svc := NewOrchestratorService(instRepo, evtRepo, &fakeContextRepo{}, &fakeCapRepo{}, nil)
	return svc, instRepo, evtRepo
}

func TestOrchestratorStartWorkflowWithIdempotencyReusesInstance(t *testing.T) {
	svc, instances, _ := newOrchestratorForTest()
	ctx := context.Background()

	first, reused, err := svc.StartWorkflowWithIdempotency(ctx, "tenant-1", "wf-1", "wv-1", "conv-1", "start-1")
	if err != nil || reused {
		t.Fatalf("first start: instance=%#v reused=%v err=%v", first, reused, err)
	}
	second, reused, err := svc.StartWorkflowWithIdempotency(ctx, "tenant-1", "wf-1", "wv-1", "conv-1", "start-1")
	if err != nil || !reused || second.ID != first.ID || instances.createCalls != 1 {
		t.Fatalf("retry start: instance=%#v reused=%v calls=%d err=%v", second, reused, instances.createCalls, err)
	}
}

func TestOrchestratorProposeEventWithIdempotencyReusesEvent(t *testing.T) {
	svc, _, events := newOrchestratorForTest()
	ctx := context.Background()
	created, err := svc.StartWorkflow(ctx, "tenant-1", "wf-1", "wv-1", "conv-1")
	if err != nil {
		t.Fatalf("start workflow: %v", err)
	}

	first, reused, err := svc.ProposeEventWithIdempotency(ctx, "tenant-1", created.ID, "doctor.specialty.selected", map[string]any{"specialty": "cardiology"}, "conv-1", "event-1")
	if err != nil || reused {
		t.Fatalf("first proposal: event=%#v reused=%v err=%v", first, reused, err)
	}
	second, reused, err := svc.ProposeEventWithIdempotency(ctx, "tenant-1", created.ID, "doctor.specialty.selected", map[string]any{"specialty": "cardiology"}, "conv-1", "event-1")
	if err != nil || !reused || second.ID != first.ID || len(events.events) != 1 {
		t.Fatalf("retry proposal: event=%#v reused=%v events=%d err=%v", second, reused, len(events.events), err)
	}
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
type fakeEngineWorkflowRepo struct {
	defs map[string]*engine.WorkflowDefinition
}

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
	sync      func(*engine.WorkflowInstance)
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
	if f.sync != nil {
		f.sync(&cp)
	}
	return nil
}

// fakeEngineEventRepo records engine events + idempotency.
type fakeEngineEventRepo struct {
	processed map[string]bool
	appended  []*engine.Event
}

func (f *fakeEngineEventRepo) Append(_ context.Context, event *engine.Event) error {
	f.appended = append(f.appended, event)
	return nil
}
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
	persistInstRepo := &fakeInstanceRepo{createStatus: entities.WorkflowInstanceCreated}
	instRepo := &fakeEngineInstanceRepo{instances: map[string]*engine.WorkflowInstance{
		"inst-1": {
			ID:                "inst-1",
			TenantID:          "tenant-1",
			WorkflowID:        "wf",
			ProjectID:         "proj",
			WorkflowVersionID: "wv-1",
			Status:            engine.InstanceCreated,
			Version:           0,
		},
	}}
	instRepo.sync = func(updated *engine.WorkflowInstance) {
		if persisted, ok := persistInstRepo.instances[updated.ID]; ok {
			persisted.Version = updated.Version
		}
	}
	eng := engine.NewEngine(engine.EngineRepositories{
		Projects:  fakeEngineProjectRepo{},
		Workflows: wfRepo,
		Instances: instRepo,
		Events:    &fakeEngineEventRepo{},
	})
	// The engine-backed orchestrator uses persistence fakes for the pre-checks,
	// and the in-memory engine repos for the actual transition.
	evtRepo := &fakeEventRepo{}
	svc := NewEngineOrchestratorService(persistInstRepo, evtRepo, &fakeContextRepo{}, &fakeCapRepo{}, eng, nil)
	return svc, instRepo
}

func TestOrchestratorStartEngineBackedInitializesEntryState(t *testing.T) {
	svc, engineInstRepo := engineBackedOrchestrator()

	created, err := svc.StartWorkflow(context.Background(), "tenant-1", "wf-1", "wv-1", "conv-1")
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	if created.Status != entities.WorkflowInstanceRunning {
		t.Fatalf("expected RUNNING, got %s", created.Status)
	}
	started, err := engineInstRepo.Get(context.Background(), "tenant-1", created.ID)
	if err != nil {
		t.Fatalf("get initialized instance: %v", err)
	}
	if started.CurrentStateID != "start" {
		t.Fatalf("expected entry state start, got %q", started.CurrentStateID)
	}
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

func capabilityGateOrchestrator() (*OrchestratorService, *fakeEngineInstanceRepo, *fakeEngineEventRepo, *fakeCapabilityEvidenceRepo) {
	const capabilityName = "padel.availability.read"
	capRepo := &fakeCapRepo{caps: map[string]*entities.Capability{
		capabilityName: {
			ID: "cap-padel", TenantID: "tenant-1", Name: capabilityName,
			ProviderType: entities.ProviderTypeMCP,
			ProviderID:   sql.NullString{String: "padel-provider-mock", Valid: true},
			ProviderTool: sql.NullString{String: "padel.cek_available", Valid: true},
		},
	}}
	definition := &engine.WorkflowDefinition{
		Slug: "wf", ProjectID: "proj", Name: "Padel", SchemaVersion: 1,
		Status: engine.WorkflowPublished, EntryNodeID: "s1",
		Nodes: []engine.WorkflowNode{
			{ID: "s1", Kind: engine.NodeKindState, Name: "Check availability", Capabilities: []string{capabilityName}},
			{ID: "end", Kind: engine.NodeKindEnd, Name: "Done"},
		},
		Transitions: []engine.TransitionDefinition{{ID: "t1", SourceStateID: "s1", Event: "availability.confirmed", TargetStateID: "end", Priority: 1}},
	}
	wfRepo := &fakeEngineWorkflowRepo{defs: map[string]*engine.WorkflowDefinition{"proj/wf": definition}}
	engineInstances := &fakeEngineInstanceRepo{instances: map[string]*engine.WorkflowInstance{
		"inst-1": {
			ID: "inst-1", TenantID: "tenant-1", ProjectID: "proj", WorkflowID: "wf",
			WorkflowVersionID: "wv-1", Status: engine.InstanceRunning, CurrentStateID: "s1", Version: 1,
		},
	}}
	engineEvents := &fakeEngineEventRepo{}
	eng := engine.NewEngine(engine.EngineRepositories{
		Projects: fakeEngineProjectRepo{}, Workflows: wfRepo, Instances: engineInstances, Events: engineEvents,
	})
	persisted := &fakeInstanceRepo{instances: map[string]*entities.WorkflowInstance{
		"inst-1": {ID: "inst-1", TenantID: "tenant-1", WorkflowID: "wf", WorkflowVersionID: "wv-1", Status: entities.WorkflowInstanceRunning, Version: 1},
	}}
	engineInstances.sync = func(updated *engine.WorkflowInstance) {
		if persistedInstance, ok := persisted.instances[updated.ID]; ok {
			persistedInstance.Status = entities.WorkflowInstanceStatus(updated.Status)
			persistedInstance.Version = updated.Version
		}
	}
	evidence := &fakeCapabilityEvidenceRepo{}
	svc := NewEngineOrchestratorService(persisted, &fakeEventRepo{}, &fakeContextRepo{}, capRepo, eng, nil, evidence)
	return svc, engineInstances, engineEvents, evidence
}

func TestOrchestratorMCPGateRejectsMissingFailedAndStaleEvidence(t *testing.T) {
	for _, tc := range []struct {
		name   string
		state  string
		status entities.CapabilityEvidenceStatus
	}{
		{name: "missing", state: "", status: ""},
		{name: "failed", state: "s1", status: entities.CapabilityEvidenceFailed},
		{name: "stale state", state: "previous-state", status: entities.CapabilityEvidenceSucceeded},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc, engineInstances, engineEvents, evidence := capabilityGateOrchestrator()
			if tc.state != "" {
				_, err := evidence.Upsert(context.Background(), repositories.CapabilityEvidenceInput{
					TenantID: "tenant-1", ProjectID: "proj", WorkflowInstanceID: "inst-1", StateID: tc.state,
					CapabilityID: "cap-padel", CapabilityName: "padel.availability.read",
					ProviderServer: "padel-provider-mock", ProviderTool: "padel.cek_available", IdempotencyKey: tc.name,
					Status: tc.status, Result: []byte(`{"available":true}`),
				})
				if err != nil {
					t.Fatalf("seed evidence: %v", err)
				}
			}

			_, err := svc.ProposeEvent(context.Background(), "tenant-1", "inst-1", "availability.confirmed", nil, "corr-1")
			if err == nil || !strings.Contains(err.Error(), "capability requirement not satisfied: padel.availability.read") {
				t.Fatalf("expected deterministic evidence gate error, got %v", err)
			}
			current, getErr := engineInstances.Get(context.Background(), "tenant-1", "inst-1")
			if getErr != nil {
				t.Fatalf("read engine instance: %v", getErr)
			}
			if current.CurrentStateID != "s1" || len(engineEvents.appended) != 0 {
				t.Fatalf("rejected transition mutated runtime: state=%s events=%d", current.CurrentStateID, len(engineEvents.appended))
			}
		})
	}
}

func TestOrchestratorMCPGateAcceptsEvidenceAndIsTenantScoped(t *testing.T) {
	svc, engineInstances, engineEvents, evidence := capabilityGateOrchestrator()
	_, err := evidence.Upsert(context.Background(), repositories.CapabilityEvidenceInput{
		TenantID: "tenant-1", ProjectID: "proj", WorkflowInstanceID: "inst-1", StateID: "s1",
		CapabilityID: "cap-padel", CapabilityName: "padel.availability.read",
		ProviderServer: "padel-provider-mock", ProviderTool: "padel.cek_available", IdempotencyKey: "availability-1",
		Status: entities.CapabilityEvidenceSucceeded, Result: []byte(`{"available":true}`),
	})
	if err != nil {
		t.Fatalf("seed success evidence: %v", err)
	}

	if _, err := svc.ProposeEvent(context.Background(), "tenant-2", "inst-1", "availability.confirmed", nil, "corr-2"); err == nil {
		t.Fatal("expected tenant-isolated instance lookup to reject cross-tenant access")
	}
	if _, err := svc.ProposeEvent(context.Background(), "tenant-1", "inst-1", "availability.confirmed", nil, "corr-1"); err != nil {
		t.Fatalf("expected matching tenant evidence to satisfy gate: %v", err)
	}
	current, err := engineInstances.Get(context.Background(), "tenant-1", "inst-1")
	if err != nil {
		t.Fatalf("read transitioned instance: %v", err)
	}
	if current.CurrentStateID != "end" || len(engineEvents.appended) != 1 {
		t.Fatalf("expected one accepted transition, state=%s events=%d", current.CurrentStateID, len(engineEvents.appended))
	}
	if current.Status != engine.InstanceCompleted {
		t.Fatalf("expected transitioned instance to be COMPLETED, got %s", current.Status)
	}
}

func TestOrchestratorMCPGateRequiresProjectBindingWhenConfigured(t *testing.T) {
	svc, _, _, evidence := capabilityGateOrchestrator()
	svc.SetProjectCapabilityMCPBindings(&fakeProjectCapabilityMCPBindingRepository{})
	_, err := evidence.Upsert(context.Background(), repositories.CapabilityEvidenceInput{
		TenantID: "tenant-1", ProjectID: "proj", WorkflowInstanceID: "inst-1", StateID: "s1",
		CapabilityID: "cap-padel", CapabilityName: "padel.availability.read",
		ProviderServer: "legacy-alias", ProviderTool: "legacy-tool", IdempotencyKey: "binding-required",
		Status: entities.CapabilityEvidenceSucceeded, Result: []byte(`{"available":true}`),
	})
	if err != nil {
		t.Fatalf("seed success evidence: %v", err)
	}
	if _, err := svc.ProposeEvent(context.Background(), "tenant-1", "inst-1", "availability.confirmed", nil, "corr-1"); err == nil || !strings.Contains(err.Error(), "capability requirement not satisfied") {
		t.Fatalf("expected missing project binding to reject transition, got %v", err)
	}
}

var _ repositories.IInstanceRepository = (*fakeInstanceRepo)(nil)
var _ repositories.IEventRepository = (*fakeEventRepo)(nil)
var _ repositories.ICapabilityEvidenceRepository = (*fakeCapabilityEvidenceRepo)(nil)
