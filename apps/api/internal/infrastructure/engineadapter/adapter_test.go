package engineadapter

import (
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/irosadie/open-state/api/internal/domain/engine"
	"github.com/irosadie/open-state/api/internal/domain/entities"
)

func TestDefinitionFromJSON(t *testing.T) {
	raw := json.RawMessage(`{
		"slug":"order-food","name":"Order Makanan","schemaVersion":1,
		"status":"PUBLISHED","projectId":"proj","entryNodeId":"n_start",
		"nodes":[{"id":"n_start","kind":"START","name":"START"}],
		"transitions":[{"id":"t0","sourceStateId":"n_start","event":"order.started","targetStateId":"n_select","priority":1}],
		"policy":{"interruptible":"USER_REQUESTED","priority":10},
		"triggers":[{"event":"order.started","source":"intent"}]
	}`)
	def, err := definitionFromJSON(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if def.Slug != "order-food" || def.ProjectID != "proj" {
		t.Errorf("unexpected def: %+v", def)
	}
	if len(def.Nodes) != 1 || def.Nodes[0].Kind != engine.NodeKindStart {
		t.Errorf("nodes not parsed: %+v", def.Nodes)
	}
	if len(def.Transitions) != 1 || def.Transitions[0].Event != "order.started" {
		t.Errorf("transitions not parsed: %+v", def.Transitions)
	}
}

func TestDefinitionFromJSONEmpty(t *testing.T) {
	if _, err := definitionFromJSON(nil); err == nil {
		t.Fatal("expected error for empty definition")
	}
}

func TestDefinitionFromJSONInvalid(t *testing.T) {
	if _, err := definitionFromJSON(json.RawMessage(`{"slug":`)); err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestToEngineInstance(t *testing.T) {
	stateID := "state-1"
	e := &entities.WorkflowInstance{
		ID:                     "inst-1",
		TenantID:               "tenant-1",
		WorkflowID:             "wf-1",
		WorkflowVersionID:      "wv-1",
		Status:                 entities.WorkflowInstanceRunning,
		Version:                3,
		CurrentStateInstanceID: &stateID,
		CorrelationKey:         sql.NullString{String: "conv-1", Valid: true},
	}
	eng := toEngineInstance(e)
	if eng.ID != "inst-1" || eng.CurrentStateID != "state-1" {
		t.Errorf("unexpected engine instance: %+v", eng)
	}
	if eng.ConversationID != "conv-1" {
		t.Errorf("expected conversation id conv-1, got %q", eng.ConversationID)
	}
	if eng.Version != 3 || eng.Status != engine.InstanceRunning {
		t.Errorf("unexpected version/status: %+v", eng)
	}
}
