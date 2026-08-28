package rag

import "context"

// Redactor masks/omits PII from text before it is returned to any client
// (PRD 90). It is a port so operators can supply their own redaction policy
// without changing the context compiler (PRD 169).
type Redactor interface {
	// Redact returns a PII-masked version of input.
	Redact(ctx context.Context, input string) (string, error)
}
