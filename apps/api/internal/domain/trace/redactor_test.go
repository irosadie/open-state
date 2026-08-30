package trace

import "testing"

func TestSanitizeAttributesRedactsSensitiveKeysAndValues(t *testing.T) {
	got := SanitizeAttributes(map[string]any{
		"reason_code": "GUARD_FAILED",
		"raw_prompt":  "customer secret prompt",
		"provider": map[string]any{
			"api_key": "sk-live-value",
			"summary": "contact user@example.com",
		},
	})

	if got["reason_code"] != "GUARD_FAILED" {
		t.Fatalf("safe reason code was not retained: %#v", got)
	}
	if got["raw_prompt"] != RedactedMarker {
		t.Fatalf("raw prompt was not redacted: %#v", got)
	}
	nested, ok := got["provider"].(map[string]any)
	if !ok || nested["api_key"] != RedactedMarker {
		t.Fatalf("nested credential was not redacted: %#v", got)
	}
	if nested["summary"] == "contact user@example.com" {
		t.Fatal("sensitive summary value was retained")
	}
}

func TestSanitizeValueRedactsNestedRAGDocuments(t *testing.T) {
	got := SanitizeValue(map[string]any{
		"retrieved_documents": []any{"private document"},
		"result_count":        2,
	})
	values := got.(map[string]any)
	if values["retrieved_documents"] != RedactedMarker {
		t.Fatalf("retrieved documents were not redacted: %#v", values)
	}
	if values["result_count"] != 2 {
		t.Fatalf("safe result count changed: %#v", values)
	}
}
