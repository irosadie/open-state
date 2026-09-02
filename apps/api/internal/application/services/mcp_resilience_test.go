package services

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	domaincap "github.com/irosadie/open-state/api/internal/domain/capability"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	domainsvc "github.com/irosadie/open-state/api/internal/domain/services"
)

type resilienceProvider struct {
	calls atomic.Int32
	err   error
}

func (p *resilienceProvider) InvokeTool(context.Context, *entities.MCPConnection, *entities.MCPDiscoveredTool, map[string]any, time.Duration) (domainsvc.MCPToolCallResult, error) {
	p.calls.Add(1)
	if p.err != nil {
		return domainsvc.MCPToolCallResult{}, p.err
	}
	return domainsvc.MCPToolCallResult{Data: map[string]any{"ok": true}}, nil
}

func testResilienceConnection() *entities.MCPConnection {
	return &entities.MCPConnection{ID: "connection-1", TenantID: "tenant-1", ProjectID: "project-1", Alias: "provider", TimeoutMS: 1000, MaxConcurrency: 1, RateLimitPerSecond: 100, RateLimitBurst: 1, RetryMax: 2, CircuitFailureThreshold: 1, CircuitRecoverySeconds: 60}
}

func testResilienceTool() *entities.MCPDiscoveredTool {
	return &entities.MCPDiscoveredTool{Name: "check_available"}
}

func TestMCPResilientProviderRetriesOnlyIdempotentCalls(t *testing.T) {
	provider := &resilienceProvider{err: domaincap.NewCapabilityError(domaincap.ErrorKindUnavailable, "provider.down", "down")}
	resilient := NewMCPResilientProvider(provider, nil, nil, nil)
	ctx := domainsvc.WithMCPCallOptions(context.Background(), domainsvc.MCPCallOptions{Idempotent: true})
	_, _ = resilient.InvokeTool(ctx, testResilienceConnection(), testResilienceTool(), nil, time.Second)
	if got := provider.calls.Load(); got != 3 {
		t.Fatalf("idempotent attempts = %d, want 3", got)
	}
}

func TestMCPResilientProviderOpensCircuitAfterFailure(t *testing.T) {
	provider := &resilienceProvider{err: domaincap.NewCapabilityError(domaincap.ErrorKindUnavailable, "provider.down", "down")}
	resilient := NewMCPResilientProvider(provider, nil, nil, nil)
	connection := testResilienceConnection()
	connection.RetryMax = 0
	_, _ = resilient.InvokeTool(context.Background(), connection, testResilienceTool(), nil, time.Second)
	_, err := resilient.InvokeTool(context.Background(), connection, testResilienceTool(), nil, time.Second)
	if err == nil {
		t.Fatal("circuit-open call unexpectedly succeeded")
	}
	if got := provider.calls.Load(); got != 1 {
		t.Fatalf("provider calls after circuit open = %d, want 1", got)
	}
}

func TestMCPResilientProviderEnforcesRateLimit(t *testing.T) {
	provider := &resilienceProvider{}
	resilient := NewMCPResilientProvider(provider, nil, nil, nil)
	connection := testResilienceConnection()
	_, _ = resilient.InvokeTool(context.Background(), connection, testResilienceTool(), nil, time.Second)
	_, err := resilient.InvokeTool(context.Background(), connection, testResilienceTool(), nil, time.Second)
	if err == nil {
		t.Fatal("second call bypassed rate limit")
	}
	if got := provider.calls.Load(); got != 1 {
		t.Fatalf("provider calls = %d, want 1", got)
	}
}
