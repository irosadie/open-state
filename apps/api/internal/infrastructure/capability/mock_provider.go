// Package capability implements the concrete capability providers for OpenState.
// This package holds infrastructure-level providers (mock/sandbox now; MCP/HTTP
// later). It implements the domain CapabilityProvider port.
package capability

import (
	"context"
	"time"

	domaincap "github.com/irosadie/open-state/api/internal/domain/capability"
)

// MockProvider is the default sandbox provider (PRD §2064). It echoes the
// invocation and returns a normalized result flagged FromMock.
type MockProvider struct{}

// Invoke implements domaincap.CapabilityProvider. It always succeeds in mock
// mode, returning a deterministic result so downstream logic is testable.
func (MockProvider) Invoke(_ context.Context, inv domaincap.Invocation) (domaincap.InvocationResult, error) {
	data := map[string]any{
		"capability": inv.Name,
		"mock":       true,
		"echo":       inv.Payload,
		"action_id":  inv.ActionID,
	}
	event := "capability.success"
	return domaincap.InvocationResult{
		Data:            data,
		FromMock:        true,
		Duration:        time.Millisecond * 5,
		CapabilityEvent: &event,
	}, nil
}

// MockProviderResolver maps every resolved capability to the MockProvider,
// serving as the default fallback when no real provider is bound (PRD §2064).
type MockProviderResolver struct{}

// ResolveProvider implements domaincap.ProviderResolver.
func (MockProviderResolver) ResolveProvider(_ *domaincap.ResolvedCapability) domaincap.CapabilityProvider {
	return MockProvider{}
}
