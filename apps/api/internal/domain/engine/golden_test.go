package engine

import (
	"context"
	"fmt"
	"testing"
)

// ---------------------------------------------------------------------------
// Golden conversation tests (PRD 125) — AI-behavior regression.
//
// A golden fixture replays user utterances through the engine and asserts the
// resolved intent / current state after each turn. Intent classification is
// stubbed to the fixture's expected intent (PRD 170), so the tests assert the
// deterministic state-machine outcome, never LLM quality. They run in CI with
// no external service.
// ---------------------------------------------------------------------------

// goldenTurn is one user utterance in a conversation: the event to emit (with
// optional payload) and the state the workflow must be in afterwards.
type goldenTurn struct {
	utterance string
	event     string
	payload   map[string]any
	expected  string // expected current-state id after this turn
}

// goldenFixture describes one end-to-end conversation against a workflow.
type goldenFixture struct {
	name       string
	def        *WorkflowDefinition
	entryEvent string
	turns      []goldenTurn
}

// runGolden replays a fixture against a fresh in-memory engine, asserting each
// turn's resolved state. On mismatch it fails with expected-vs-actual.
func runGolden(t *testing.T, fx goldenFixture) {
	t.Helper()
	repos := newFakeRepos()
	if err := repos.Workflows.Save(context.Background(), fx.def); err != nil {
		t.Fatalf("save workflow: %v", err)
	}
	eng := NewEngine(repos)

	inst, err := eng.StartWorkflow(context.Background(), "demo", fx.def.ProjectID, "conv", fx.def, fx.entryEvent)
	if err != nil {
		t.Fatalf("start: %v", err)
	}

	// Advance through the turns; each maps an utterance to an event + expected state.
	for i, turn := range fx.turns {
		next, _, err := eng.ProcessEvent(context.Background(), "demo", inst.ID, &Event{
			ID:      fmt.Sprintf("e%d", i),
			Type:    turn.event,
			Source:  SourceUser,
			Payload: turn.payload,
		})
		if err != nil {
			t.Fatalf("turn %d (%s): event %s: %v", i, turn.utterance, turn.event, err)
		}
		if next.CurrentStateID != turn.expected {
			t.Errorf("turn %d (%s): expected state %q, got %q", i, turn.utterance, turn.expected, next.CurrentStateID)
		}
	}
}

// --- Fixture: PADEL court booking -----------------------------------------

func goldenPadel() goldenFixture {
	return goldenFixture{
		name:       "padel-court-booking",
		def:        padelDef(),
		entryEvent: "workflow.started",
		turns: []goldenTurn{
			{utterance: "user starts booking", event: "workflow.started", expected: "select_time"},
			{utterance: "user picks a date/time", event: "datetime.selected", payload: map[string]any{"booking.date": "2026-08-27", "booking.time": "19:00"}, expected: "check_stock"},
			{utterance: "slot is available", event: "slot.available", payload: map[string]any{"slot.available": true}, expected: "confirm"},
			{utterance: "user confirms", event: "confirm.requested", expected: "done"},
		},
	}
}

func TestGoldenPadel(t *testing.T) { runGolden(t, goldenPadel()) }

// --- Fixture: ORDER FOOD ---------------------------------------------------

func orderFoodDef() *WorkflowDefinition {
	return &WorkflowDefinition{
		Slug:          "order-food",
		ProjectID:     "retail",
		Name:          "Order Makanan",
		SchemaVersion: 1,
		Status:        WorkflowPublished,
		EntryNodeID:   "n_start",
		Nodes: []WorkflowNode{
			{ID: "n_start", Kind: NodeKindStart, Name: "START"},
			{ID: "n_select_product", Kind: NodeKindState, Name: "SELECT_PRODUCT", RequiredContext: []string{"order.items"}},
			{ID: "n_check_stock", Kind: NodeKindDecision, Name: "CHECK_STOCK", RequiredContext: []string{"product.sku"}},
			{ID: "n_collect_customer", Kind: NodeKindState, Name: "COLLECT_CUSTOMER", RequiredContext: []string{"customer.name", "customer.address"}},
			{ID: "n_payment", Kind: NodeKindState, Name: "PAYMENT", RequiredContext: []string{"order.total"}},
			{ID: "n_order_confirmed", Kind: NodeKindEnd, Name: "ORDER_CONFIRMED", IsTerminal: true},
		},
		Transitions: []TransitionDefinition{
			{ID: "t0", SourceStateID: "n_start", Event: "order.started", TargetStateID: "n_select_product", Priority: 1},
			{ID: "t1", SourceStateID: "n_select_product", Event: "product.requested", TargetStateID: "n_check_stock", Priority: 1,
				Guards: []GuardGroup{{Logic: "AND", Conditions: []GuardCondition{{Field: "product.sku", Operator: OpExists}}}}},
			{ID: "t2", SourceStateID: "n_check_stock", Event: "product.in_stock", TargetStateID: "n_collect_customer", Priority: 1,
				Guards: []GuardGroup{{Logic: "AND", Conditions: []GuardCondition{{Field: "product.in_stock", Operator: OpEq, Value: true}}}}},
			{ID: "t3", SourceStateID: "n_collect_customer", Event: "customer.ready", TargetStateID: "n_payment", Priority: 1,
				Guards: []GuardGroup{{Logic: "AND", Conditions: []GuardCondition{{Field: "customer.name", Operator: OpExists}, {Field: "customer.address", Operator: OpExists}}}}},
			{ID: "t4", SourceStateID: "n_payment", Event: "payment.success", TargetStateID: "n_order_confirmed", Priority: 1,
				Guards: []GuardGroup{{Logic: "AND", Conditions: []GuardCondition{{Field: "payment.status", Operator: OpEq, Value: "success"}}}}},
		},
		Policy: WorkflowPolicy{Interruptible: "USER_REQUESTED", Priority: 10},
	}
}

func goldenOrderFood() goldenFixture {
	return goldenFixture{
		name:       "order-food",
		def:        orderFoodDef(),
		entryEvent: "order.started",
		turns: []goldenTurn{
			{utterance: "start order", event: "order.started", expected: "n_select_product"},
			{utterance: "user requests a product", event: "product.requested", payload: map[string]any{"product.sku": "latte"}, expected: "n_check_stock"},
			{utterance: "product is in stock", event: "product.in_stock", payload: map[string]any{"product.in_stock": true}, expected: "n_collect_customer"},
			{utterance: "customer details ready", event: "customer.ready", payload: map[string]any{"customer.name": "Rina", "customer.address": "Jakarta"}, expected: "n_payment"},
			{utterance: "payment succeeds", event: "payment.success", payload: map[string]any{"payment.status": "success"}, expected: "n_order_confirmed"},
		},
	}
}

func TestGoldenOrderFood(t *testing.T) { runGolden(t, goldenOrderFood()) }

// --- Fixture: ORDER DOCTOR -------------------------------------------------

func orderDoctorDef() *WorkflowDefinition {
	return &WorkflowDefinition{
		Slug:          "order-doctor",
		ProjectID:     "health",
		Name:          "Order Dokter",
		SchemaVersion: 1,
		Status:        WorkflowPublished,
		EntryNodeID:   "n_start",
		Nodes: []WorkflowNode{
			{ID: "n_start", Kind: NodeKindStart, Name: "START"},
			{ID: "n_select_specialty", Kind: NodeKindState, Name: "SELECT_SPECIALTY", RequiredContext: []string{"doctor.specialty"}},
			{ID: "n_check_availability", Kind: NodeKindDecision, Name: "CHECK_AVAILABILITY", RequiredContext: []string{"doctor.id"}},
			{ID: "n_select_time_slot", Kind: NodeKindState, Name: "SELECT_TIME_SLOT", RequiredContext: []string{"booking.time_start"}},
			{ID: "n_booking_confirmed", Kind: NodeKindEnd, Name: "BOOKING_CONFIRMED", IsTerminal: true},
		},
		Transitions: []TransitionDefinition{
			{ID: "t0", SourceStateID: "n_start", Event: "consultation.requested", TargetStateID: "n_select_specialty", Priority: 1},
			{ID: "t1", SourceStateID: "n_select_specialty", Event: "doctor.requested", TargetStateID: "n_check_availability", Priority: 1,
				Guards: []GuardGroup{{Logic: "AND", Conditions: []GuardCondition{{Field: "doctor.specialty", Operator: OpExists}}}}},
			{ID: "t2", SourceStateID: "n_check_availability", Event: "doctor.available", TargetStateID: "n_select_time_slot", Priority: 1,
				Guards: []GuardGroup{{Logic: "AND", Conditions: []GuardCondition{{Field: "schedule.available", Operator: OpEq, Value: true}}}}},
			{ID: "t3", SourceStateID: "n_select_time_slot", Event: "slot.confirmed", TargetStateID: "n_booking_confirmed", Priority: 1,
				Guards: []GuardGroup{{Logic: "AND", Conditions: []GuardCondition{{Field: "booking.time_start", Operator: OpExists}}}}},
		},
		Policy: WorkflowPolicy{Interruptible: "USER_REQUESTED", Priority: 10},
	}
}

func goldenOrderDoctor() goldenFixture {
	return goldenFixture{
		name:       "order-doctor",
		def:        orderDoctorDef(),
		entryEvent: "consultation.requested",
		turns: []goldenTurn{
			{utterance: "start consultation", event: "consultation.requested", expected: "n_select_specialty"},
			{utterance: "user asks for a doctor", event: "doctor.requested", payload: map[string]any{"doctor.specialty": "gizi"}, expected: "n_check_availability"},
			{utterance: "doctor available", event: "doctor.available", payload: map[string]any{"schedule.available": true}, expected: "n_select_time_slot"},
			{utterance: "user picks a time slot", event: "slot.confirmed", payload: map[string]any{"booking.time_start": "09:00"}, expected: "n_booking_confirmed"},
		},
	}
}

func TestGoldenOrderDoctor(t *testing.T) { runGolden(t, goldenOrderDoctor()) }
