package capability

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/irosadie/open-state/api/internal/domain/entities"
)

// fakeProvider returns a fixed result or error; counts invocations.
type fakeProvider struct {
	err      error
	result   InvocationResult
	invoked  atomic.Int32
	failures atomic.Int32
}

func (f *fakeProvider) Invoke(_ context.Context, _ Invocation) (InvocationResult, error) {
	f.invoked.Add(1)
	if f.err != nil {
		f.failures.Add(1)
		return InvocationResult{}, f.err
	}
	return f.result, nil
}

// stubProviderResolver always returns the given provider.
type stubProviderResolver struct{ p CapabilityProvider }

func (s stubProviderResolver) ResolveProvider(*ResolvedCapability) CapabilityProvider { return s.p }

type stubSchemaValidator struct{ fail bool }

func (s stubSchemaValidator) Validate(_ map[string]any, _ []byte) error {
	if s.fail {
		return errors.New("invalid")
	}
	return nil
}

type allowRateLimiter struct{}

func (allowRateLimiter) Allow(_ context.Context, _ string) (bool, error) { return true, nil }

// denyRateLimiter rejects every request, simulating an exceeded rate limit.
type denyRateLimiter struct{}

func (denyRateLimiter) Allow(_ context.Context, _ string) (bool, error) { return false, nil }

// recordingRateLimiter allows and records the last key passed to Allow.
type recordingRateLimiter struct{ lastKey string }

func (r *recordingRateLimiter) Allow(_ context.Context, key string) (bool, error) {
	r.lastKey = key
	return true, nil
}

func baseInvocation() Invocation {
	return Invocation{
		TenantID: "t", ProjectID: "p", WorkflowID: "wf", StateID: "st",
		ActionID: "a1", WorkflowInstanceID: "wfi1", Name: "payment.create",
		Payload: map[string]any{},
		Policy:  CapabilityPolicy{Timeout: 0, MaxRetry: 0},
	}
}

func TestInvokerSuccess(t *testing.T) {
	repo := &fakeCapabilityRepo{}
	pid := "mcp"
	_, _ = repo.Create(context.Background(), "t", "payment.create", nil, "MCP", &pid, nil, []byte(`{}`), nil, 1, nil)
	prov := &fakeProvider{result: InvocationResult{Data: map[string]any{"ok": true}}}
	inv := NewCapabilityInvoker(
		NewCapabilityResolver(repo), stubProviderResolver{prov},
		stubSchemaValidator{}, allowRateLimiter{}, NewInMemoryIdempotencyStore(),
	)
	res, err := inv.Execute(context.Background(), baseInvocation())
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}
	if res.Data["ok"] != true {
		t.Errorf("unexpected result: %+v", res.Data)
	}
	if prov.invoked.Load() != 1 {
		t.Errorf("expected 1 invocation, got %d", prov.invoked.Load())
	}
}

func TestInvokerSchemaValidationFails(t *testing.T) {
	repo := &fakeCapabilityRepo{}
	pid := "mcp"
	_, _ = repo.Create(context.Background(), "t", "payment.create", nil, "MCP", &pid, nil, []byte(`{}`), nil, 1, nil)
	prov := &fakeProvider{}
	inv := NewCapabilityInvoker(
		NewCapabilityResolver(repo), stubProviderResolver{prov},
		stubSchemaValidator{fail: true}, allowRateLimiter{}, NewInMemoryIdempotencyStore(),
	)
	_, err := inv.Execute(context.Background(), baseInvocation())
	var ce *CapabilityError
	if !errors.As(err, &ce) || ce.Kind != ErrorKindValidation {
		t.Fatalf("expected validation error, got %v", err)
	}
	if prov.invoked.Load() != 0 {
		t.Error("provider must NOT be invoked on schema failure")
	}
}

func TestInvokerRateLimitedNoInvoke(t *testing.T) {
	repo := &fakeCapabilityRepo{}
	pid := "mcp"
	_, _ = repo.Create(context.Background(), "t", "payment.create", nil, "MCP", &pid, nil, []byte(`{}`), nil, 1, nil)
	prov := &fakeProvider{}
	inv := NewCapabilityInvoker(
		NewCapabilityResolver(repo), stubProviderResolver{prov},
		stubSchemaValidator{}, denyRateLimiter{}, NewInMemoryIdempotencyStore(),
	)
	_, err := inv.Execute(context.Background(), baseInvocation())
	var ce *CapabilityError
	if !errors.As(err, &ce) || ce.Kind != ErrorKindRateLimited {
		t.Fatalf("expected rate_limited error, got %v", err)
	}
	if ce.Code != "capability.rate_limited" {
		t.Errorf("expected code capability.rate_limited, got %q", ce.Code)
	}
	if prov.invoked.Load() != 0 {
		t.Error("provider must NOT be invoked when rate limited")
	}
}

func TestInvokerRateLimitedKeyScoped(t *testing.T) {
	repo := &fakeCapabilityRepo{}
	pid := "mcp"
	_, _ = repo.Create(context.Background(), "t", "payment.create", nil, "MCP", &pid, nil, []byte(`{}`), nil, 1, nil)
	prov := &fakeProvider{}

	// Capture the key passed to the rate limiter to verify tenant+capability scope.
	rl := &recordingRateLimiter{}
	inv := NewCapabilityInvoker(
		NewCapabilityResolver(repo), stubProviderResolver{prov},
		stubSchemaValidator{}, rl, NewInMemoryIdempotencyStore(),
	)
	invOC := baseInvocation()
	invOC.CapabilityID = "cap-123"
	invOC.TenantID = "tenant-9"
	if _, err := inv.Execute(context.Background(), invOC); err != nil {
		t.Fatalf("execute error: %v", err)
	}
	if rl.lastKey != "tenant:tenant-9:capability:cap-123" {
		t.Errorf("expected scoped key, got %q", rl.lastKey)
	}
}

func TestInvokerDeniedNoInvoke(t *testing.T) {
	repo := &fakeCapabilityRepo{}
	pid := "mcp"
	_, _ = repo.Create(context.Background(), "t", "payment.create", nil, "MCP", &pid, nil, nil, nil, 1, nil)
	repo.bindings = []entities.CapabilityBinding{
		{ScopeType: entities.BindingScopeState, ScopeID: "st", Permission: entities.BindingPermissionDeny},
	}
	prov := &fakeProvider{}
	inv := NewCapabilityInvoker(
		NewCapabilityResolver(repo), stubProviderResolver{prov},
		stubSchemaValidator{}, allowRateLimiter{}, NewInMemoryIdempotencyStore(),
	)
	_, err := inv.Execute(context.Background(), baseInvocation())
	if err == nil {
		t.Fatal("expected denial")
	}
	if prov.invoked.Load() != 0 {
		t.Error("provider must NOT be invoked on denial")
	}
}

func TestInvokerRetryOnRetryable(t *testing.T) {
	repo := &fakeCapabilityRepo{}
	pid := "mcp"
	_, _ = repo.Create(context.Background(), "t", "payment.create", nil, "MCP", &pid, nil, nil, nil, 1, nil)
	prov := &fakeProvider{err: NewCapabilityError(ErrorKindTimeout, "capability.timeout", "timeout")}
	inv := NewCapabilityInvoker(
		NewCapabilityResolver(repo), stubProviderResolver{prov},
		stubSchemaValidator{}, allowRateLimiter{}, NewInMemoryIdempotencyStore(),
	)
	invInvocation := baseInvocation()
	invInvocation.Policy = CapabilityPolicy{MaxRetry: 2, Retryable: []ErrorKind{ErrorKindTimeout}}
	_, err := inv.Execute(context.Background(), invInvocation)
	if err == nil {
		t.Fatal("expected error after retries exhausted")
	}
	if prov.invoked.Load() != 3 { // 1 initial + 2 retries
		t.Errorf("expected 3 invocations, got %d", prov.invoked.Load())
	}
}

func TestInvokerNonRetryableNoRetry(t *testing.T) {
	repo := &fakeCapabilityRepo{}
	pid := "mcp"
	_, _ = repo.Create(context.Background(), "t", "payment.create", nil, "MCP", &pid, nil, nil, nil, 1, nil)
	prov := &fakeProvider{err: NewCapabilityError(ErrorKindUnauthorized, "capability.unauthorized", "denied")}
	inv := NewCapabilityInvoker(
		NewCapabilityResolver(repo), stubProviderResolver{prov},
		stubSchemaValidator{}, allowRateLimiter{}, NewInMemoryIdempotencyStore(),
	)
	invInvocation := baseInvocation()
	invInvocation.Policy = CapabilityPolicy{MaxRetry: 5}
	_, err := inv.Execute(context.Background(), invInvocation)
	if err == nil {
		t.Fatal("expected error")
	}
	if prov.invoked.Load() != 1 {
		t.Errorf("non-retryable must not retry, got %d invocations", prov.invoked.Load())
	}
}

func TestInvokerIdempotency(t *testing.T) {
	repo := &fakeCapabilityRepo{}
	pid := "mcp"
	_, _ = repo.Create(context.Background(), "t", "payment.create", nil, "MCP", &pid, nil, nil, nil, 1, nil)
	prov := &fakeProvider{result: InvocationResult{Data: map[string]any{"ok": true}}}
	store := NewInMemoryIdempotencyStore()
	inv := NewCapabilityInvoker(
		NewCapabilityResolver(repo), stubProviderResolver{prov},
		stubSchemaValidator{}, allowRateLimiter{}, store,
	)
	// first run invokes
	_, _ = inv.Execute(context.Background(), baseInvocation())
	// duplicate run with same key returns cached, no provider call
	_, err := inv.Execute(context.Background(), baseInvocation())
	if err != nil {
		t.Fatalf("duplicate error: %v", err)
	}
	if prov.invoked.Load() != 1 {
		t.Errorf("expected 1 provider invocation (idempotent), got %d", prov.invoked.Load())
	}
}
