package dtos

// CompiledContext is the normalized, PII-redacted per-turn context returned to an
// LLM/RAG client (PRD 22, 24, 43.2, 90). Available/missing reflect the runtime
// context; memory and workflow are kept separate so long-lived customer data never
// leaks into workflow runtime state.
type CompiledContext struct {
	// Available is the merged runtime context available for the turn.
	Available map[string]any `json:"available"`
	// Missing lists required keys absent for the current state.
	Missing []string `json:"missing"`
	// Memory holds long-lived customer/user memory (PRD §24).
	Memory map[string]any `json:"memory"`
	// Workflow holds transient workflow runtime data (PRD §24).
	Workflow map[string]any `json:"workflow"`
	// Retrieval holds RAG knowledge retrieved for the turn, if any (PRD 171).
	Retrieval []map[string]any `json:"retrieval,omitempty"`
	// Redacted reports whether PII redaction was applied (PRD 90).
	Redacted bool `json:"redacted"`
}
