package engine

import (
	"crypto/rand"
	"encoding/hex"
)

// newID generates a random hex id (36 chars, prefixed) for entities.
func newID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// extremely unlikely; fall back to timestamp-based id
		return fallbackID()
	}
	return "id_" + hex.EncodeToString(b)
}

func fallbackID() string {
	// stable fallback; not cryptographically random but never blocks
	return "id_fallback"
}

// defSlugVersion produces a stable version key for a definition.
// In production this is the immutable workflow_version_id; here we derive from slug.
func defSlugVersion(def *WorkflowDefinition) string {
	return def.Slug + "@" + "v1"
}
