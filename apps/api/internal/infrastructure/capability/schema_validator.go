// Package capability provides the infrastructure-level capability providers
// and sandbox helpers for the capability admin API (Epic #4).
package capability

import (
	"encoding/json"
	"fmt"
)

// JSONSchemaValidator implements domaincap.InputSchemaValidator with a minimal
// JSON-Schema subset (type: object, properties, required) sufficient for
// sandbox test-invocation (PRD §62). It intentionally does not implement the
// full JSON Schema spec; it guards the common admin-testing cases.
type JSONSchemaValidator struct{}

// Validate checks that payload conforms to the given JSON Schema subset.
func (JSONSchemaValidator) Validate(payload map[string]any, schema []byte) error {
	if len(schema) == 0 || string(schema) == "{}" {
		return nil
	}
	var def struct {
		Type       string                     `json:"type"`
		Required   []string                   `json:"required"`
		Properties map[string]json.RawMessage `json:"properties"`
	}
	if err := json.Unmarshal(schema, &def); err != nil {
		return fmt.Errorf("invalid input schema: %w", err)
	}
	for _, field := range def.Required {
		if _, ok := payload[field]; !ok {
			return fmt.Errorf("missing required field %q", field)
		}
	}
	if def.Type != "" && def.Type != "object" && def.Type != "null" {
		return fmt.Errorf("payload must be an object")
	}
	return nil
}
