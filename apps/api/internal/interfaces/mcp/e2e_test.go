package mcp

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/irosadie/open-state/api/internal/domain/engine"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	domain "github.com/irosadie/open-state/go-shared/domain"
)

// ---------------------------------------------------------------------------
// E2E test (PRD 170): a deterministic LLM/MCP mock drives the real MCP tool
// handlers (resolve_intent, start_workflow, propose_event) through the engine to
// a state transition, and the resulting state is asserted via the repository.
// No real LLM or database is involved — everything is in-memory and repeatable.
// ---------------------------------------------------------------------------

// --- In-memory fakes for the engine ports (test-local) ----------------------

type e2eWorkflowRepo struct{ defs map[string]*engine.WorkflowDefinition }

func (r *e2eWorkflowRepo) GetBySlug(_ context.Context, _, projectID, slug string) (*engine.WorkflowDefinition, error) {
	d, ok := r.defs[projectID+"/"+slug]
	if !ok {
		return nil, domain.NewNotFound("workflow not found")
	}
	return d, nil
}
func (r *e2eWorkflowRepo) Save(_ context.Context, d *engine.WorkflowDefinition) error {
	if r.defs == nil {
		r.defs = map[string]*engine.WorkflowDefinition{}
	}
	r.defs[d.ProjectID+"/"+d.Slug] = d
	return nil
}

type e2eInstanceRepo struct{ insts map[string]*engine.WorkflowInstance }

func (r *e2eInstanceRepo) Create(_ context.Context, i *engine.WorkflowInstance) error {
	if r.insts == nil {
		r.insts = map[string]*engine.WorkflowInstance{}
	}
	cp := *i
	r.insts[i.ID] = &cp
	return nil
}
func (r *e2eInstanceRepo) Get(_ context.Context, _, id string) (*engine.WorkflowInstance, error) {
	i, ok := r.insts[id]
	if !ok {
		return nil, domain.NewNotFound("instance not found")
	}
	cp := *i
	return &cp, nil
}
func (r *e2eInstanceRepo) UpdateWithVersion(_ context.Context, i *engine.WorkflowInstance, expected int) error {
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

type e2eEventRepo struct{ processed map[string]bool }

func (r *e2eEventRepo) Append(context.Context, *engine.Event) error { return nil }
func (r *e2eEventRepo) IsProcessed(_ context.Context, _, key string) (bool, error) {
	return r.processed[key], nil
}
func (r *e2eEventRepo) MarkProcessed(_ context.Context, _, key, _ string) error {
	if r.processed == nil {
		r.processed = map[string]bool{}
	}
	r.processed[key] = true
	return nil
}

type e2eProjectRepo struct{}

func (e2eProjectRepo) Get(context.Context, string, string) (*engine.Project, error) {
	return &engine.Project{ID: "proj"}, nil
}
func (e2eProjectRepo) Save(context.Context, *engine.Project) error { return nil }

// --- Mock orchestrator bridging the MCP port to the in-memory engine -------

// mockOrchestrator implements OrchestratorPort on top of the engine + fakes.
type mockOrchestrator struct {
	eng         *engine.Engine
	workflows   *e2eWorkflowRepo
	instances   *e2eInstanceRepo
	def         *engine.WorkflowDefinition
	convToInst  map[string]string // conversationID -> instanceID
	lastStateID string
}

func newMockOrchestrator(def *engine.WorkflowDefinition) *mockOrchestrator {
	wfRepo := &e2eWorkflowRepo{}
	_ = wfRepo.Save(context.Background(), def)
	instRepo := &e2eInstanceRepo{}
	eng := engine.NewEngine(engine.EngineRepositories{
		Projects:  e2eProjectRepo{},
		Workflows: wfRepo,
		Instances: instRepo,
		Events:    &e2eEventRepo{},
	})
	return &mockOrchestrator{eng: eng, workflows: wfRepo, instances: instRepo, def: def, convToInst: map[string]string{}}
}

func (o *mockOrchestrator) StartWorkflow(_ context.Context, _ string, workflowID, _ string, correlation string) (*entities.WorkflowInstance, error) {
	inst, err := o.eng.StartWorkflow(context.Background(), "demo", o.def.ProjectID, correlation, o.def, o.def.Triggers[0].Event)
	if err != nil {
		return nil, err
	}
	o.convToInst[correlation] = inst.ID
	return toEntitiesInstance(inst), nil
}

func (o *mockOrchestrator) ProposeEvent(_ context.Context, _ string, instanceID, eventType string, payload map[string]any, _ string) (*entities.Event, error) {
	next, _, err := o.eng.ProcessEvent(context.Background(), "demo", instanceID, &engine.Event{
		ID:      "e2e-" + eventType,
		Type:    eventType,
		Source:  engine.SourceUser,
		Payload: payload,
	})
	if err != nil {
		return nil, err
	}
	o.lastStateID = next.CurrentStateID
	return &entities.Event{ID: "e2e-" + eventType, Type: eventType}, nil
}

func (o *mockOrchestrator) GetCurrentState(_ context.Context, _, instanceID string) (*entities.WorkflowInstance, *entities.StateInstance, error) {
	inst, err := o.instances.Get(context.Background(), "demo", instanceID)
	if err != nil {
		return nil, nil, err
	}
	stateID := inst.CurrentStateID
	return toEntitiesInstance(inst), &entities.StateInstance{StateID: &stateID, Status: entities.StateInstanceActive}, nil
}

func (o *mockOrchestrator) GetActiveWorkflow(_ context.Context, _, conversationID string) (*entities.WorkflowInstance, error) {
	id, ok := o.convToInst[conversationID]
	if !ok {
		return nil, domain.NewNotFound("no active workflow")
	}
	inst, err := o.instances.Get(context.Background(), "demo", id)
	if err != nil {
		return nil, err
	}
	return toEntitiesInstance(inst), nil
}

func (o *mockOrchestrator) SuspendWorkflow(context.Context, string, string) (*entities.WorkflowInstance, error) { return nil, nil }
func (o *mockOrchestrator) ResumeWorkflow(context.Context, string, string) (*entities.WorkflowInstance, error)  { return nil, nil }
func (o *mockOrchestrator) CancelWorkflow(context.Context, string, string) (*entities.WorkflowInstance, error)  { return nil, nil }
func (o *mockOrchestrator) ListInstances(context.Context, string) ([]entities.WorkflowInstance, error)          { return nil, nil }
func (o *mockOrchestrator) ListHistory(context.Context, string, string) ([]entities.Event, error)               { return nil, nil }
func (o *mockOrchestrator) ReplayWorkflow(context.Context, string, string) (map[string]any, *entities.Event, error) {
	return map[string]any{}, nil, nil
}
func (o *mockOrchestrator) ListAllowedCapabilities(context.Context, string, entities.BindingScopeType, string) ([]entities.Capability, error) {
	return nil, nil
}

func (o *mockOrchestrator) GetAllowedTransitions(ctx context.Context, _, instanceID string) ([]engine.TransitionDefinition, error) {
	return o.eng.AllowedTransitions(ctx, "demo", instanceID)
}

func toEntitiesInstance(i *engine.WorkflowInstance) *entities.WorkflowInstance {
	stateID := i.CurrentStateID
	return &entities.WorkflowInstance{
		ID:                     i.ID,
		TenantID:               i.TenantID,
		WorkflowID:             i.WorkflowID,
		Status:                 entities.WorkflowInstanceStatus(i.Status),
		Version:                i.Version,
		CurrentStateInstanceID: &stateID,
	}
}

// --- Intent mock -----------------------------------------------------------

type e2eIntentResolver struct{ def *engine.WorkflowDefinition }

func (e2eIntentResolver) ListIntents(context.Context, string, string) ([]entities.Workflow, error) {
	return []entities.Workflow{{ID: "ORDER_FOOD", Slug: "order-food", Name: "Order Makanan"}}, nil
}
func (e2eIntentResolver) ResolveIntent(_ context.Context, _, _, intentID string) (*entities.Workflow, error) {
	return &entities.Workflow{ID: intentID, Slug: "order-food", Name: "Order Makanan", Status: entities.WorkflowPublished}, nil
}

// --- The E2E test ----------------------------------------------------------

func TestE2EFullPath(t *testing.T) {
	def := e2eOrderFoodDef()
	orch := newMockOrchestrator(def)
	deps := Dependencies{
		IntentResolver: e2eIntentResolver{def: def},
		Orchestrator:   orch,
	}
	ts := server.NewTestStreamableHTTPServer(NewServer(deps))
	defer ts.Close()

	c, err := client.NewStreamableHttpClient(ts.URL + "/mcp")
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	if _, err := c.Initialize(context.Background(), mcp.InitializeRequest{}); err != nil {
		t.Fatalf("initialize: %v", err)
	}

	// 1. resolve_intent → the mock LLM resolves the user intent to the workflow.
	if res, err := c.CallTool(context.Background(), mcp.CallToolRequest{Params: mcp.CallToolParams{
		Name: "resolve_intent",
		Arguments: map[string]any{"tenant": "demo", "intent": "ORDER_FOOD", "project": "retail"},
	}}); err != nil || res.IsError {
		t.Fatalf("resolve_intent failed: err=%v res=%+v", err, res)
	}

	// 2. start_workflow → begin an order-food instance.
	startRes, err := c.CallTool(context.Background(), mcp.CallToolRequest{Params: mcp.CallToolParams{
		Name:      "start_workflow",
		Arguments: map[string]any{"tenant": "demo", "workflow": "order-food", "correlation": "conv-1"},
	}})
	if err != nil || startRes.IsError {
		t.Fatalf("start_workflow failed: err=%v res=%+v", err, startRes)
	}
	instanceID := orch.convToInst["conv-1"]
	if instanceID == "" {
		t.Fatal("instance not created")
	}

	// 3. propose_event → drive the engine through product selection + stock.
	propose := func(evt string, payload map[string]any) {
		t.Helper()
		res, err := c.CallTool(context.Background(), mcp.CallToolRequest{Params: mcp.CallToolParams{
			Name: "propose_event",
			Arguments: map[string]any{"tenant": "demo", "instance": instanceID, "type": evt, "payload": payload},
		}})
		if err != nil || res.IsError {
			t.Fatalf("propose_event %s failed: err=%v res=%+v", evt, err, res)
		}
	}

	propose("order.started", nil)
	propose("product.requested", map[string]any{"product.sku": "latte"})
	propose("product.in_stock", map[string]any{"product.in_stock": true})
	propose("customer.ready", map[string]any{"customer.name": "Rina", "customer.address": "Jakarta"})
	propose("payment.success", map[string]any{"payment.status": "success"})

	// 4. Assert persisted state via the repository (the fake instance repo).
	inst, err := orch.instances.Get(context.Background(), "demo", instanceID)
	if err != nil {
		t.Fatalf("read persisted instance: %v", err)
	}
	if inst.CurrentStateID != "n_order_confirmed" {
		t.Errorf("expected terminal n_order_confirmed, got %q", inst.CurrentStateID)
	}
}

// e2eOrderFoodDef mirrors the seeded order-food workflow (engine format).
func e2eOrderFoodDef() *engine.WorkflowDefinition {
	return &engine.WorkflowDefinition{
		Slug:          "order-food",
		ProjectID:     "retail",
		Name:          "Order Makanan",
		SchemaVersion: 1,
		Status:        engine.WorkflowPublished,
		EntryNodeID:   "n_start",
		Nodes: []engine.WorkflowNode{
			{ID: "n_start", Kind: engine.NodeKindStart, Name: "START"},
			{ID: "n_select_product", Kind: engine.NodeKindState, Name: "SELECT_PRODUCT"},
			{ID: "n_check_stock", Kind: engine.NodeKindDecision, Name: "CHECK_STOCK"},
			{ID: "n_collect_customer", Kind: engine.NodeKindState, Name: "COLLECT_CUSTOMER"},
			{ID: "n_payment", Kind: engine.NodeKindState, Name: "PAYMENT"},
			{ID: "n_order_confirmed", Kind: engine.NodeKindEnd, Name: "ORDER_CONFIRMED", IsTerminal: true},
		},
		Transitions: []engine.TransitionDefinition{
			{ID: "t0", SourceStateID: "n_start", Event: "order.started", TargetStateID: "n_select_product", Priority: 1},
			{ID: "t1", SourceStateID: "n_select_product", Event: "product.requested", TargetStateID: "n_check_stock", Priority: 1,
				Guards: []engine.GuardGroup{{Logic: "AND", Conditions: []engine.GuardCondition{{Field: "product.sku", Operator: engine.OpExists}}}}},
			{ID: "t2", SourceStateID: "n_check_stock", Event: "product.in_stock", TargetStateID: "n_collect_customer", Priority: 1,
				Guards: []engine.GuardGroup{{Logic: "AND", Conditions: []engine.GuardCondition{{Field: "product.in_stock", Operator: engine.OpEq, Value: true}}}}},
			{ID: "t3", SourceStateID: "n_collect_customer", Event: "customer.ready", TargetStateID: "n_payment", Priority: 1,
				Guards: []engine.GuardGroup{{Logic: "AND", Conditions: []engine.GuardCondition{{Field: "customer.name", Operator: engine.OpExists}, {Field: "customer.address", Operator: engine.OpExists}}}}},
			{ID: "t4", SourceStateID: "n_payment", Event: "payment.success", TargetStateID: "n_order_confirmed", Priority: 1,
				Guards: []engine.GuardGroup{{Logic: "AND", Conditions: []engine.GuardCondition{{Field: "payment.status", Operator: engine.OpEq, Value: "success"}}}}},
		},
		Triggers: []engine.WorkflowTrigger{{Event: "order.started", Source: "intent"}},
		Policy:   engine.WorkflowPolicy{Interruptible: "USER_REQUESTED", Priority: 10},
	}
}
