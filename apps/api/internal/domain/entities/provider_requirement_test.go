package entities

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestProviderRequirementValidateRequiresSafeMapping(t *testing.T) {
	req := ProviderRequirement{Capability: "padel.availability.read", Purpose: "Check available courts", Required: true}
	if err := req.Validate(); err == nil {
		t.Fatal("expected missing provider mapping to fail validation")
	}

	req.ProviderServer = "padel-provider"
	req.Tool = "padel.check_available"
	if err := req.Validate(); err != nil {
		t.Fatalf("expected valid provider requirement: %v", err)
	}
}

func TestProviderRequirementSerializationDoesNotContainSecretsOrEndpoints(t *testing.T) {
	req := ProviderRequirement{
		Capability: "padel.availability.read", ProviderServer: "padel-provider", Tool: "padel.check_available",
		Purpose: "Check available courts", Required: true,
	}
	raw, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal requirement: %v", err)
	}
	text := string(raw)
	if strings.Contains(text, "Bearer") || strings.Contains(text, "secret") || strings.Contains(text, "://") {
		t.Fatalf("unsafe provider metadata serialized: %s", text)
	}
}
