package services

import (
	"encoding/json"
	"testing"
)

func TestValidateWorkflowDefinition(t *testing.T) {
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{
			name:  "valid graph",
			input: `{"nodes":[{"id":"start","kind":"START"},{"id":"end","kind":"END"}],"transitions":[{"id":"t1","sourceStateId":"start","targetStateId":"end"}]}`,
			valid: true,
		},
		{
			name:  "missing start",
			input: `{"nodes":[{"id":"end","kind":"END"}],"transitions":[]}`,
		},
		{
			name:  "duplicate start",
			input: `{"nodes":[{"id":"start-1","kind":"START"},{"id":"start-2","kind":"START"},{"id":"end","kind":"END"}],"transitions":[]}`,
		},
		{
			name:  "invalid transition endpoint",
			input: `{"nodes":[{"id":"start","kind":"START"},{"id":"end","kind":"END"}],"transitions":[{"id":"t1","sourceStateId":"start","targetStateId":"missing"}]}`,
		},
		{
			name:  "no terminal path",
			input: `{"nodes":[{"id":"start","kind":"START"},{"id":"end","kind":"END"}],"transitions":[]}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateWorkflowDefinition([]byte(test.input))
			if (err == nil) != test.valid {
				t.Fatalf("validation result mismatch: err=%v", err)
			}
			if err != nil && len(err.Details) == 0 {
				t.Fatal("expected structured validation details")
			}
		})
	}
}

func TestCompareWorkflowDefinitionsDeterministic(t *testing.T) {
	base := []byte(`{"nodes":[{"id":"start","kind":"START"},{"id":"old","kind":"STATE"}],"transitions":[{"id":"old-transition","sourceStateId":"start","targetStateId":"old"}]}`)
	target := []byte(`{"nodes":[{"id":"start","kind":"START"},{"id":"new","kind":"STATE"}],"transitions":[{"id":"new-transition","sourceStateId":"start","targetStateId":"new"}]}`)
	diff, err := compareWorkflowDefinitions("wf-1", 1, 2, base, target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diff.Nodes.Added) != 1 || diff.Nodes.Added[0].ID != "new" {
		t.Fatalf("unexpected added nodes: %+v", diff.Nodes.Added)
	}
	if len(diff.Nodes.Removed) != 1 || diff.Nodes.Removed[0].ID != "old" {
		t.Fatalf("unexpected removed nodes: %+v", diff.Nodes.Removed)
	}
	if len(diff.Transitions.Added) != 1 || len(diff.Transitions.Removed) != 1 {
		t.Fatalf("unexpected transition diff: %+v", diff.Transitions)
	}
	if _, err := json.Marshal(diff); err != nil {
		t.Fatalf("diff should be serializable: %v", err)
	}
}

func TestCompareWorkflowDefinitionsChangedFields(t *testing.T) {
	base := []byte(`{"nodes":[{"id":"start","kind":"START","name":"Start"}],"transitions":[]}`)
	target := []byte(`{"nodes":[{"id":"start","kind":"START","name":"Renamed"}],"transitions":[]}`)
	diff, err := compareWorkflowDefinitions("wf-1", 1, 2, base, target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diff.Nodes.Changed) != 1 || len(diff.Nodes.Changed[0].ChangedFields) != 1 || diff.Nodes.Changed[0].ChangedFields[0] != "name" {
		t.Fatalf("unexpected changed fields: %+v", diff.Nodes.Changed)
	}
}

func TestCompareWorkflowDefinitionsIgnoresObjectKeyOrder(t *testing.T) {
	base := []byte(`{"nodes":[{"id":"start","kind":"START","policy":{"timeout":10,"retry":false}}],"transitions":[]}`)
	target := []byte(`{"nodes":[{"id":"start","kind":"START","policy":{"retry":false,"timeout":10}}],"transitions":[]}`)
	diff, err := compareWorkflowDefinitions("wf-1", 1, 2, base, target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diff.Nodes.Changed) != 0 {
		t.Fatalf("expected no semantic change, got %+v", diff.Nodes.Changed)
	}
}
