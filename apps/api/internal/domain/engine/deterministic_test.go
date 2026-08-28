package engine

import (
	"context"
	"errors"
	"testing"

	"github.com/irosadie/open-state/go-shared/domain"
)

// ---------------------------------------------------------------------------
// Deterministic runtime tests (PRD 126) — no LLM, pure engine inputs.
// These exercise `event → guard → transition` directly and assert exact
// deterministic outcomes, so they can run in CI without any external service.
// ---------------------------------------------------------------------------

// evalCtx is a tiny helper returning a context map.
func evalCtx(kv map[string]any) map[string]any { return kv }

// ---- Guard operator coverage ----

func TestGuardOperators(t *testing.T) {
	tests := []struct {
		name string
		cond GuardCondition
		ctx  map[string]any
		want bool
	}{
		// == numeric
		{"eq int match", GuardCondition{Field: "x", Operator: OpEq, Value: 5}, evalCtx(map[string]any{"x": 5}), true},
		{"eq int mismatch", GuardCondition{Field: "x", Operator: OpEq, Value: 5}, evalCtx(map[string]any{"x": 6}), false},
		{"eq string-num", GuardCondition{Field: "x", Operator: OpEq, Value: "100"}, evalCtx(map[string]any{"x": 100}), true},
		// !=
		{"neq differ", GuardCondition{Field: "x", Operator: OpNeq, Value: 5}, evalCtx(map[string]any{"x": 6}), true},
		{"neq equal", GuardCondition{Field: "x", Operator: OpNeq, Value: 5}, evalCtx(map[string]any{"x": 5}), false},
		// >
		{"gt true", GuardCondition{Field: "x", Operator: OpGt, Value: 10}, evalCtx(map[string]any{"x": 15}), true},
		{"gt false", GuardCondition{Field: "x", Operator: OpGt, Value: 10}, evalCtx(map[string]any{"x": 5}), false},
		// >=
		{"gte equal", GuardCondition{Field: "x", Operator: OpGte, Value: 10}, evalCtx(map[string]any{"x": 10}), true},
		{"gte below", GuardCondition{Field: "x", Operator: OpGte, Value: 10}, evalCtx(map[string]any{"x": 9}), false},
		// <
		{"lt true", GuardCondition{Field: "x", Operator: OpLt, Value: 10}, evalCtx(map[string]any{"x": 4}), true},
		{"lt false", GuardCondition{Field: "x", Operator: OpLt, Value: 10}, evalCtx(map[string]any{"x": 12}), false},
		// <=
		{"lte equal", GuardCondition{Field: "x", Operator: OpLte, Value: 10}, evalCtx(map[string]any{"x": 10}), true},
		{"lte above", GuardCondition{Field: "x", Operator: OpLte, Value: 10}, evalCtx(map[string]any{"x": 11}), false},
		// IN
		{"in true", GuardCondition{Field: "role", Operator: OpIn, Value: []any{"admin", "user"}}, evalCtx(map[string]any{"role": "user"}), true},
		{"in false", GuardCondition{Field: "role", Operator: OpIn, Value: []any{"admin", "user"}}, evalCtx(map[string]any{"role": "guest"}), false},
		// NOT_IN
		{"not_in true", GuardCondition{Field: "role", Operator: OpNotIn, Value: []any{"admin", "user"}}, evalCtx(map[string]any{"role": "guest"}), true},
		{"not_in false", GuardCondition{Field: "role", Operator: OpNotIn, Value: []any{"admin", "user"}}, evalCtx(map[string]any{"role": "user"}), false},
		// EXISTS
		{"exists true", GuardCondition{Field: "user.id", Operator: OpExists}, evalCtx(map[string]any{"user.id": "u1"}), true},
		{"exists false", GuardCondition{Field: "user.id", Operator: OpExists}, evalCtx(map[string]any{}), false},
		// NOT_EXISTS
		{"not_exists true", GuardCondition{Field: "user.id", Operator: OpNotExists}, evalCtx(map[string]any{}), true},
		{"not_exists false", GuardCondition{Field: "user.id", Operator: OpNotExists}, evalCtx(map[string]any{"user.id": "u1"}), false},
		// nested dot-path resolution
		{"nested exists", GuardCondition{Field: "payment.status", Operator: OpEq, Value: "success"}, evalCtx(map[string]any{"payment": map[string]any{"status": "success"}}), true},
		{"missing field treated false", GuardCondition{Field: "nonexistent", Operator: OpEq, Value: 1}, evalCtx(map[string]any{}), false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := EvaluateGuardGroup(GuardGroup{Logic: "AND", Conditions: []GuardCondition{tc.cond}}, tc.ctx)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

// ---- AND / OR grouping ----

func TestGuardGrouping(t *testing.T) {
	ctx := evalCtx(map[string]any{"a": 1, "b": 2, "c": 3})

	// AND: all must pass
	andPass := GuardGroup{Logic: "AND", Conditions: []GuardCondition{
		{Field: "a", Operator: OpEq, Value: 1},
		{Field: "b", Operator: OpEq, Value: 2},
	}}
	if ok, _ := EvaluateGuardGroup(andPass, ctx); !ok {
		t.Error("AND group with all passing should be true")
	}
	andFail := GuardGroup{Logic: "AND", Conditions: []GuardCondition{
		{Field: "a", Operator: OpEq, Value: 1},
		{Field: "b", Operator: OpEq, Value: 99},
	}}
	if ok, _ := EvaluateGuardGroup(andFail, ctx); ok {
		t.Error("AND group with one failing should be false")
	}

	// OR: any passes
	orPass := GuardGroup{Logic: "OR", Conditions: []GuardCondition{
		{Field: "a", Operator: OpEq, Value: 99},
		{Field: "b", Operator: OpEq, Value: 2},
	}}
	if ok, _ := EvaluateGuardGroup(orPass, ctx); !ok {
		t.Error("OR group with one passing should be true")
	}
	orFail := GuardGroup{Logic: "OR", Conditions: []GuardCondition{
		{Field: "a", Operator: OpEq, Value: 99},
		{Field: "b", Operator: OpEq, Value: 99},
	}}
	if ok, _ := EvaluateGuardGroup(orFail, ctx); ok {
		t.Error("OR group with all failing should be false")
	}

	// Multiple groups: all must pass (AND across groups)
	groups := []GuardGroup{
		{Logic: "AND", Conditions: []GuardCondition{{Field: "a", Operator: OpEq, Value: 1}}},
		{Logic: "AND", Conditions: []GuardCondition{{Field: "b", Operator: OpEq, Value: 2}}},
	}
	if ok, _ := EvaluateGuards(groups, ctx); !ok {
		t.Error("multiple passing groups should be true")
	}
	failing := []GuardGroup{
		{Logic: "AND", Conditions: []GuardCondition{{Field: "a", Operator: OpEq, Value: 1}}},
		{Logic: "AND", Conditions: []GuardCondition{{Field: "b", Operator: OpEq, Value: 99}}},
	}
	if ok, _ := EvaluateGuards(failing, ctx); ok {
		t.Error("one failing group should make guards false")
	}
}

// ---- Priority ordering (PRD §34) ----

// priorityDef builds a workflow with multiple transitions from a single state
// sharing the same event but with different guards + priorities.
func priorityDef() *WorkflowDefinition {
	return &WorkflowDefinition{
		Slug:          "priority",
		ProjectID:     "p",
		Name:          "Priority",
		SchemaVersion: 1,
		Status:        WorkflowPublished,
		EntryNodeID:   "start",
		Nodes: []WorkflowNode{
			{ID: "start", Kind: NodeKindStart, Name: "START"},
			{ID: "s1", Kind: NodeKindState, Name: "S1"},
			{ID: "s2", Kind: NodeKindState, Name: "S2"},
			{ID: "s3", Kind: NodeKindState, Name: "S3"},
		},
		Transitions: []TransitionDefinition{
			{ID: "t0", SourceStateID: "start", Event: "go", TargetStateID: "s1", Priority: 1},
			// Three transitions on "go" from s1, guarded by tier.
			{ID: "t_high", SourceStateID: "s1", Event: "go", TargetStateID: "s3", Priority: 1,
				Guards: []GuardGroup{{Logic: "AND", Conditions: []GuardCondition{{Field: "tier", Operator: OpEq, Value: "high"}}}}},
			{ID: "t_med", SourceStateID: "s1", Event: "go", TargetStateID: "s2", Priority: 5,
				Guards: []GuardGroup{{Logic: "AND", Conditions: []GuardCondition{{Field: "tier", Operator: OpEq, Value: "med"}}}}},
			{ID: "t_low", SourceStateID: "s1", Event: "go", TargetStateID: "s1", Priority: 10,
				Guards: []GuardGroup{{Logic: "AND", Conditions: []GuardCondition{{Field: "tier", Operator: OpEq, Value: "low"}}}}},
		},
		Policy: WorkflowPolicy{Interruptible: "NEVER", Priority: 1},
	}
}

func TestPriorityOrdering(t *testing.T) {
	repos := newFakeRepos()
	def := priorityDef()
	_ = repos.Workflows.Save(context.Background(), def)
	eng := NewEngine(repos)
	inst, _ := eng.StartWorkflow(context.Background(), "t", "p", "c", def, "go")
	_, _, _ = eng.ProcessEvent(context.Background(), "t", inst.ID, &Event{ID: "e0", Type: "go", Source: SourceSystem})

	// All three transitions are candidates for "go" from s1, but only the guard
	// matching the tier passes. tier=med selects t_med (priority 5) -> s2.
	inst2, _, err := eng.ProcessEvent(context.Background(), "t", inst.ID, &Event{ID: "e1", Type: "go", Source: SourceUser, Payload: map[string]any{"tier": "med"}})
	if err != nil {
		t.Fatalf("process event: %v", err)
	}
	if inst2.CurrentStateID != "s2" {
		t.Errorf("expected s2 (priority 5), got %q", inst2.CurrentStateID)
	}
}

func TestPriorityHighestPassingWins(t *testing.T) {
	repos := newFakeRepos()
	def := priorityDef()
	_ = repos.Workflows.Save(context.Background(), def)
	eng := NewEngine(repos)
	inst, _ := eng.StartWorkflow(context.Background(), "t", "p", "c", def, "go")
	_, _, _ = eng.ProcessEvent(context.Background(), "t", inst.ID, &Event{ID: "e0", Type: "go", Source: SourceSystem})

	// Multiple guards pass when tier matches two? Build a def where both high
	// and med guards pass, assert the lower (higher-priority) number wins.
	def2 := priorityDef()
	// force t_high and t_med to both match "gold": override conditions.
	def2.Transitions[1].Guards = []GuardGroup{{Logic: "AND", Conditions: []GuardCondition{{Field: "tier", Operator: OpEq, Value: "gold"}}}}
	def2.Transitions[2].Guards = []GuardGroup{{Logic: "AND", Conditions: []GuardCondition{{Field: "tier", Operator: OpEq, Value: "gold"}}}}
	_ = repos.Workflows.Save(context.Background(), def2)
	inst2, _ := eng.StartWorkflow(context.Background(), "t", "p", "c2", def2, "go")
	_, _, _ = eng.ProcessEvent(context.Background(), "t", inst2.ID, &Event{ID: "e0", Type: "go", Source: SourceSystem})
	inst3, _, err := eng.ProcessEvent(context.Background(), "t", inst2.ID, &Event{ID: "e1", Type: "go", Source: SourceUser, Payload: map[string]any{"tier": "gold"}})
	if err != nil {
		t.Fatalf("process event: %v", err)
	}
	// t_high (priority 1) must win over t_med (priority 5) -> s3
	if inst3.CurrentStateID != "s3" {
		t.Errorf("expected s3 (priority 1 wins), got %q", inst3.CurrentStateID)
	}
}

// ---- Rejection: disallowed event ----

func TestRejectDisallowedEvent(t *testing.T) {
	repos := newFakeRepos()
	def := padelDef()
	_ = repos.Workflows.Save(context.Background(), def)
	eng := NewEngine(repos)
	inst, _ := eng.StartWorkflow(context.Background(), "t", "project-padel", "c", def, "workflow.started")
	// Move to select_time, then capture the current (post-move) state.
	inst, _, err := eng.ProcessEvent(context.Background(), "t", inst.ID, &Event{ID: "e0", Type: "workflow.started", Source: SourceSystem})
	if err != nil {
		t.Fatalf("workflow.started: %v", err)
	}

	// "payment.success" is not valid from select_time.
	_, _, err = eng.ProcessEvent(context.Background(), "t", inst.ID, &Event{ID: "e1", Type: "payment.success", Source: SourceUser})
	if err == nil {
		t.Fatal("expected error for disallowed event")
	}
	var de *domain.DomainError
	if !errors.As(err, &de) || de.Code != domain.ErrConflict {
		t.Fatalf("expected CONFLICT error, got %v", err)
	}
	// State must be unchanged.
	after, _ := repos.Instances.Get(context.Background(), "t", inst.ID)
	if after.CurrentStateID != inst.CurrentStateID {
		t.Errorf("state changed on rejection: %q -> %q", inst.CurrentStateID, after.CurrentStateID)
	}
}

func TestRejectNoPassingGuard(t *testing.T) {
	repos := newFakeRepos()
	def := padelDef()
	_ = repos.Workflows.Save(context.Background(), def)
	eng := NewEngine(repos)
	inst, _ := eng.StartWorkflow(context.Background(), "t", "project-padel", "c", def, "workflow.started")
	_, _, _ = eng.ProcessEvent(context.Background(), "t", inst.ID, &Event{ID: "e0", Type: "workflow.started", Source: SourceSystem})
	_, _, _ = eng.ProcessEvent(context.Background(), "t", inst.ID, &Event{ID: "e1", Type: "datetime.selected", Source: SourceUser})

	// slot.available event with slot unavailable => guard fails, no transition.
	_, _, err := eng.ProcessEvent(context.Background(), "t", inst.ID, &Event{
		ID: "e2", Type: "slot.available", Source: SourceMCP, Payload: map[string]any{"slot.available": false},
	})
	if err == nil {
		t.Fatal("expected ErrGuardFailed when no transition passes guards")
	}
}
