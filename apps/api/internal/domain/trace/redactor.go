package trace

import (
	"regexp"
	"strings"

	"github.com/irosadie/open-state/api/internal/domain/entities"
)

const RedactedMarker = "[REDACTED]"

var (
	emailPattern       = regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)
	phonePattern       = regexp.MustCompile(`\+?\d[\d.\- ]{6,}\d`)
	cardPattern        = regexp.MustCompile(`\b(?:\d[ -]*?){13,16}\b`)
	secretValuePattern = regexp.MustCompile(`(?i)\b(?:bearer|basic)\s+[A-Za-z0-9._~+/=-]+`)
)

// SanitizeAttributes applies a stable key deny-list plus conservative value
// redaction recursively. The result is safe to persist and safe to serialize.
func SanitizeAttributes(input map[string]any) entities.SanitizedAttributes {
	return entities.SanitizedAttributes(sanitizeMap(input))
}

// SanitizeValue applies the same policy to persisted context projections.
func SanitizeValue(input any) any {
	return sanitizeValue(input)
}

func sanitizeMap(input map[string]any) map[string]any {
	out := make(map[string]any, len(input))
	for key, value := range input {
		if sensitiveKey(key) {
			out[key] = RedactedMarker
			continue
		}
		out[key] = sanitizeValue(value)
	}
	return out
}

func sanitizeValue(value any) any {
	switch v := value.(type) {
	case map[string]any:
		return sanitizeMap(v)
	case []any:
		out := make([]any, len(v))
		for i := range v {
			out[i] = sanitizeValue(v[i])
		}
		return out
	case string:
		return redactText(v)
	default:
		return value
	}
}

func sensitiveKey(key string) bool {
	normalized := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(key, "-", "_"), " ", "_"))
	for _, token := range []string{
		"secret", "password", "passwd", "credential", "token", "authorization", "api_key", "apikey",
		"private_key", "raw_prompt", "prompt", "raw_response", "model_response", "response_body",
		"document", "documents", "retrieved", "email", "phone", "mobile", "address", "ssn", "national_id",
	} {
		if strings.Contains(normalized, token) {
			return true
		}
	}
	return false
}

func redactText(input string) string {
	out := emailPattern.ReplaceAllString(input, RedactedMarker)
	out = phonePattern.ReplaceAllString(out, RedactedMarker)
	out = cardPattern.ReplaceAllString(out, RedactedMarker)
	out = secretValuePattern.ReplaceAllString(out, RedactedMarker)
	return out
}
