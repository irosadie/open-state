package capability

import (
	"context"
	"testing"

	domaincap "github.com/irosadie/open-state/api/internal/domain/capability"
)

func TestMockProviderReturnsMockResult(t *testing.T) {
	mp := MockProvider{}
	inv := domaincap.Invocation{Name: "payment.create", ActionID: "a1", Payload: map[string]any{"x": 1}}
	res, err := mp.Invoke(context.Background(), inv)
	if err != nil {
		t.Fatalf("mock invoke error: %v", err)
	}
	if !res.FromMock {
		t.Error("expected fromMock=true")
	}
	if res.Data["capability"] != "payment.create" {
		t.Errorf("unexpected data: %+v", res.Data)
	}
	if res.CapabilityEvent == nil || *res.CapabilityEvent != "capability.success" {
		t.Errorf("expected success event, got %v", res.CapabilityEvent)
	}
}

func TestMockProviderResolverReturnsMock(t *testing.T) {
	resolver := MockProviderResolver{}
	p := resolver.ResolveProvider(&domaincap.ResolvedCapability{})
	// must implement the port (compile-time check via invocation)
	_, err := p.Invoke(context.Background(), domaincap.Invocation{Name: "x"})
	if err != nil {
		t.Fatalf("resolved mock invoke error: %v", err)
	}
}
