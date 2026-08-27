// Package capability implements the safe, deterministic capability execution
// layer for OpenState (Epic #4, mcp-capability-execution).
//
// It is domain-pure: it has no HTTP, database, or MCP-SDK dependency (PRD §172,
// §2559). It depends only on the ICapabilityRepository port (defined in
// persistence-capabilities-policies) and on a CapabilityProvider port that a
// concrete provider (MCP, INTERNAL, HTTP) implements.
package capability

import (
	"context"
	"time"
)

// CapabilityProvider is the port through which a logical capability is
// executed. The core engine stays agnostic to the concrete MCP/HTTP/INTERNAL
// implementation; each is a CapabilityProvider implementation (PRD §172, §59).
type CapabilityProvider interface {
	// Invoke executes a capability invocation and returns a normalized result
	// or a classified error. Implementations MUST NOT expose raw provider
	// internals; errors are mapped to CapabilityError.
	Invoke(ctx context.Context, inv Invocation) (InvocationResult, error)
}

// CapabilityPolicy carries per-invocation runtime constraints (PRD §88, §160).
type CapabilityPolicy struct {
	// Timeout is the per-call deadline applied via context.WithTimeout.
	Timeout time.Duration
	// MaxRetry is the retry budget for retryable errors.
	MaxRetry int
	// Retryable lists the error kinds eligible for retry (empty = none).
	Retryable []ErrorKind
}

// Invocation is the full, authorized request to run a capability (PRD §62, §64).
type Invocation struct {
	TenantID     string
	ProjectID    string
	WorkflowID   string
	WorkflowInstanceID string
	StateID      string
	ActionID     string
	CapabilityID string // resolved logical capability id
	Name         string // logical capability name, e.g. payment.create
	Payload      map[string]any
	IdempotencyKey string // workflow_instance_id + action_id (PRD §64)
	Policy       CapabilityPolicy
}

// InvocationResult is the normalized outcome of a capability execution.
type InvocationResult struct {
	Data            map[string]any
	FromMock        bool   // true when executed via the default mock/sandbox provider
	Duration        time.Duration
	CapabilityEvent *string // e.g. "capability.success", nil if none
}
