// Package rag defines the portable knowledge-retrieval (RAG) port and the PII
// redaction port for the context compiler (PRD 171, 90, 169). The State Engine
// and context compiler depend only on these interfaces, never on a concrete RAG
// backend, keeping the platform replaceable and LLM-agnostic (PRD 170).
package rag

import "context"

// RAGProvider retrieves relevant knowledge for a query (PRD 171).
// It is a domain port: implementations (vector DB, search API, ...) are injected
// at the composition root.
type RAGProvider interface {
	// Retrieve returns relevant knowledge for the given query. It returns only
	// retrieved text/metadata — never a generated answer (PRD 170).
	Retrieve(ctx context.Context, query string) (*Retrieval, error)
}

// Retrieval is a normalized chunk of retrieved knowledge usable by the context
// compiler without leaking backend-specific types (PRD 171).
type Retrieval struct {
	// Text is the retrieved source text.
	Text string
	// Metadata carries optional, non-PII metadata (relevance, source, ...).
	Metadata map[string]any
}
