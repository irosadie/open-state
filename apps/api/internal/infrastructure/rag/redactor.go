// Package rag provides default infrastructure implementations of the RAG and PII
// redaction ports. This is a stub/simple implementation until a concrete RAG
// backend lands (PRD 171).
package rag

import (
	"context"
	"regexp"

	domainrag "github.com/irosadie/open-state/api/internal/domain/rag"
)

// StubRAGProvider is a no-op RAG provider wired at the composition root until a
// concrete backend is integrated. It returns no retrievals (PRD 171).
type StubRAGProvider struct{}

// Retrieve implements domainrag.RAGProvider. Returns an empty retrieval.
func (StubRAGProvider) Retrieve(_ context.Context, _ string) (*domainrag.Retrieval, error) {
	return &domainrag.Retrieval{Text: ""}, nil
}

// DefaultRedactor masks common PII patterns (email, phone, credit-card-like
// numbers) conservatively (PRD 90). It is configurable via a pattern set.
type DefaultRedactor struct {
	// Patterns applied in order; each is replaced by its label.
	Patterns []PIIPattern
}

// PIIPattern pairs a compiled regex with a replacement label.
type PIIPattern struct {
	Regex       *regexp.Regexp
	Replacement string
}

// DefaultPatterns returns the conservative default PII patterns (PRD 90).
func DefaultPatterns() []PIIPattern {
	return []PIIPattern{
		{Regex: regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`), Replacement: "[email redacted]"},
		// Phone: must start with + or have spaces/dashes between digit groups.
		// Excludes bare digit strings like booking IDs and ISO dates (YYYY-MM-DD).
		{Regex: regexp.MustCompile(`(?:\+\d{1,3}[\s\-]?)?\(?\d{2,4}\)?[\s\-]\d{3,4}[\s\-]\d{3,6}`), Replacement: "[phone redacted]"},
		{Regex: regexp.MustCompile(`\b(?:\d[ -]*?){13,16}\b`), Replacement: "[card redacted]"},
	}
}

// NewDefaultRedactor builds a DefaultRedactor with the default patterns.
func NewDefaultRedactor() *DefaultRedactor {
	return &DefaultRedactor{Patterns: DefaultPatterns()}
}

// Redact implements domainrag.Redactor.
func (r *DefaultRedactor) Redact(_ context.Context, input string) (string, error) {
	out := input
	for _, p := range r.Patterns {
		out = p.Regex.ReplaceAllString(out, p.Replacement)
	}
	return out, nil
}

// Ensure the implementations satisfy the ports at compile time.
var (
	_ domainrag.RAGProvider = StubRAGProvider{}
	_ domainrag.Redactor    = (*DefaultRedactor)(nil)
)
