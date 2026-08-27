package capability

import (
	"testing"
)

func TestJSONSchemaValidatorMissingRequired(t *testing.T) {
	v := JSONSchemaValidator{}
	schema := []byte(`{"type":"object","required":["amount"],"properties":{"amount":{"type":"number"}}}`)
	err := v.Validate(map[string]any{}, schema)
	if err == nil {
		t.Fatal("expected missing-required error")
	}
}

func TestJSONSchemaValidatorValid(t *testing.T) {
	v := JSONSchemaValidator{}
	schema := []byte(`{"type":"object","required":["amount"],"properties":{"amount":{"type":"number"}}}`)
	if err := v.Validate(map[string]any{"amount": 100}, schema); err != nil {
		t.Fatalf("expected valid, got %v", err)
	}
}

func TestJSONSchemaValidatorEmptySchema(t *testing.T) {
	v := JSONSchemaValidator{}
	if err := v.Validate(map[string]any{"x": 1}, []byte(`{}`)); err != nil {
		t.Fatalf("empty schema should pass, got %v", err)
	}
}
