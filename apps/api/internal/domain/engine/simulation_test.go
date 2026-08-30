package engine

import (
	"context"
	"reflect"
	"testing"
)

func TestSimulateSuccessfulPathProducesTrace(t *testing.T) {
	def := padelDef()
	def.Nodes[3].Capabilities = []string{"booking.confirm"}
	input := SimulationInput{
		TenantID:   "tenant",
		ProjectID:  "project-padel",
		Definition: def,
		InitialContext: map[string]any{
			"booking": map[string]any{"date": "2026-08-27", "time": "19:00"},
		},
		Events: []SimulationEvent{
			{Type: "workflow.started"},
			{Type: "datetime.selected"},
			{Type: "slot.available", Payload: map[string]any{"slot.available": true}},
			{Type: "confirm.requested"},
		},
	}

	result, err := NewEngine(newFakeRepos()).Simulate(context.Background(), input)
	if err != nil {
		t.Fatalf("simulate: %v", err)
	}
	if result.FinalState.ID != "done" || result.FinalStatus != InstanceCompleted {
		t.Fatalf("expected completed done state, got %+v", result)
	}
	if result.FinalContext["slot.available"] != true {
		t.Fatalf("expected final context to retain event payload, got %+v", result.FinalContext)
	}
	if len(result.Steps) != 5 {
		t.Fatalf("expected entry plus four event steps, got %d", len(result.Steps))
	}
	if result.Steps[3].SelectedTransitionID != "t3" || result.Steps[3].Outcome != SimulationTransitioned {
		t.Fatalf("expected t3 transition trace, got %+v", result.Steps[3])
	}
	if len(result.Steps[3].CapabilityRequests) != 1 || !result.Steps[3].CapabilityRequests[0].Mock {
		t.Fatalf("expected mock capability plan on confirm, got %+v", result.Steps[3].CapabilityRequests)
	}
}

func TestSimulateReportsAllCandidatesAndPriority(t *testing.T) {
	def := priorityDef()
	def.Transitions[1].Guards = []GuardGroup{{Logic: "AND", Conditions: []GuardCondition{{Field: "tier", Operator: OpEq, Value: "gold"}}}}
	def.Transitions[2].Guards = []GuardGroup{{Logic: "AND", Conditions: []GuardCondition{{Field: "tier", Operator: OpEq, Value: "gold"}}}}
	result, err := NewEngine(newFakeRepos()).Simulate(context.Background(), SimulationInput{
		TenantID:   "tenant",
		ProjectID:  "p",
		Definition: def,
		Events: []SimulationEvent{
			{Type: "go"},
			{Type: "go", Payload: map[string]any{"tier": "gold"}},
		},
	})
	if err != nil {
		t.Fatalf("simulate: %v", err)
	}
	step := result.Steps[2]
	if len(step.Candidates) != 3 {
		t.Fatalf("expected all three candidates, got %+v", step.Candidates)
	}
	if step.SelectedTransitionID != "t_high" || step.StateAfter == nil || step.StateAfter.ID != "s3" {
		t.Fatalf("expected lowest priority passing transition, got %+v", step)
	}
	if !step.Candidates[0].GuardPassed || !step.Candidates[1].GuardPassed || step.Candidates[2].GuardPassed {
		t.Fatalf("unexpected candidate guard outcomes: %+v", step.Candidates)
	}
}

func TestSimulateStopsOnGuardRejectionWithoutMutatingContext(t *testing.T) {
	def := padelDef()
	result, err := NewEngine(newFakeRepos()).Simulate(context.Background(), SimulationInput{
		TenantID:   "tenant",
		ProjectID:  "project-padel",
		Definition: def,
		Events: []SimulationEvent{
			{Type: "workflow.started"},
			{Type: "datetime.selected"},
			{Type: "slot.available", Payload: map[string]any{"slot.available": false}},
			{Type: "confirm.requested"},
		},
	})
	if err != nil {
		t.Fatalf("simulate: %v", err)
	}
	if len(result.Steps) != 4 || result.Steps[3].ErrorCode != "GUARD_FAILED" {
		t.Fatalf("expected stop at guard failure, got %+v", result.Steps)
	}
	if result.FinalState.ID != "check_stock" {
		t.Fatalf("expected state to remain check_stock, got %s", result.FinalState.ID)
	}
	if _, exists := result.FinalContext["slot.available"]; exists {
		t.Fatalf("rejected event payload must not be committed to context: %+v", result.FinalContext)
	}
}

func TestSimulateIsRepeatableAndDoesNotUseProductionRepositories(t *testing.T) {
	repos := newFakeRepos()
	def := padelDef()
	eng := NewEngine(repos)
	input := SimulationInput{
		TenantID:   "tenant",
		ProjectID:  "project-padel",
		Definition: def,
		Events:     []SimulationEvent{{Type: "workflow.started"}},
	}
	first, err := eng.Simulate(context.Background(), input)
	if err != nil {
		t.Fatalf("first simulate: %v", err)
	}
	second, err := eng.Simulate(context.Background(), input)
	if err != nil {
		t.Fatalf("second simulate: %v", err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("simulation must be repeatable:\nfirst=%+v\nsecond=%+v", first, second)
	}
	instanceRepo := repos.Instances.(*fakeInstanceRepo)
	workflowRepo := repos.Workflows.(*fakeWorkflowRepo)
	if len(instanceRepo.insts) != 0 {
		t.Fatalf("simulation wrote to production instance repository: %+v", instanceRepo.insts)
	}
	if len(workflowRepo.defs) != 0 {
		t.Fatalf("simulation wrote to production workflow repository: %+v", workflowRepo.defs)
	}
}

func TestSimulateRejectsTooManyEvents(t *testing.T) {
	events := make([]SimulationEvent, MaxSimulationEvents+1)
	_, err := NewEngine(newFakeRepos()).Simulate(context.Background(), SimulationInput{
		TenantID:   "tenant",
		ProjectID:  "project-padel",
		Definition: padelDef(),
		Events:     events,
	})
	if err == nil {
		t.Fatal("expected event limit validation error")
	}
}
