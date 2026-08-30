package services

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/irosadie/open-state/api/internal/application/dtos"
	domainengine "github.com/irosadie/open-state/api/internal/domain/engine"
)

func TestSimulationServiceRunsUnsavedDefinition(t *testing.T) {
	definition, err := json.Marshal(domainengine.WorkflowDefinition{
		Slug:        "member-registration",
		Name:        "Member registration",
		EntryNodeID: "start",
		Nodes: []domainengine.WorkflowNode{
			{ID: "start", Kind: domainengine.NodeKindStart, Name: "START"},
			{ID: "finish", Kind: domainengine.NodeKindEnd, Name: "FINISH", IsTerminal: true, Capabilities: []string{"member.notify"}},
		},
		Transitions: []domainengine.TransitionDefinition{{ID: "start-finish", SourceStateID: "start", Event: "member.registered", TargetStateID: "finish", Priority: 1}},
	})
	if err != nil {
		t.Fatalf("marshal definition: %v", err)
	}

	result, err := NewSimulationService().Simulate(context.Background(), "tenant-1", dtos.SimulateWorkflowRequest{
		Definition:     definition,
		InitialContext: map[string]json.RawMessage{"member": json.RawMessage(`{"name":"Ada"}`)},
		Events:         []dtos.SimulationEvent{{Type: "member.registered", Payload: map[string]json.RawMessage{"member.id": json.RawMessage(`"m-1"`)}}},
	})
	if err != nil {
		t.Fatalf("simulate: %v", err)
	}
	if result.FinalState.ID != "finish" || result.FinalStatus != string(domainengine.InstanceCompleted) {
		t.Fatalf("unexpected final result: %+v", result)
	}
	if len(result.Steps) != 2 || result.Steps[1].SelectedTransitionID != "start-finish" {
		t.Fatalf("unexpected trace: %+v", result.Steps)
	}
	if !result.Steps[1].CapabilityRequests[0].Mock || result.Steps[1].CapabilityRequests[0].Status != "PLANNED" {
		t.Fatalf("expected mock capability plan: %+v", result.Steps[1].CapabilityRequests)
	}
}

func TestSimulationServiceValidatesInput(t *testing.T) {
	service := NewSimulationService()
	if _, err := service.Simulate(context.Background(), "tenant-1", dtos.SimulateWorkflowRequest{}); err == nil {
		t.Fatal("expected missing definition validation")
	}

	definition, _ := json.Marshal(domainengine.WorkflowDefinition{EntryNodeID: "start", Nodes: []domainengine.WorkflowNode{{ID: "start", Kind: domainengine.NodeKindStart, Name: "START"}}})
	if _, err := service.Simulate(context.Background(), "tenant-1", dtos.SimulateWorkflowRequest{
		Definition: definition,
		Events:     []dtos.SimulationEvent{{Payload: map[string]json.RawMessage{}}},
	}); err == nil {
		t.Fatal("expected event type validation")
	}
	if _, err := service.Simulate(context.Background(), "tenant-1", dtos.SimulateWorkflowRequest{
		Definition: definition,
		Events:     []dtos.SimulationEvent{{Type: "   "}},
	}); err == nil {
		t.Fatal("expected whitespace event type validation")
	}
}
