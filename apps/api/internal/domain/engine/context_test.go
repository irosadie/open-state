package engine

import "testing"

func TestContextResolverPrecedence(t *testing.T) {
	r := NewContextResolver()
	r.Set(ScopeTenant, "merchant.name", "Cafe ABC", false)
	r.Set(ScopeWorkflow, "merchant.name", "Cafe XYZ", false) // later overrides
	r.Set(ScopeWorkflow, "booking.date", "2026-08-27", false)
	r.Set(ScopeState, "booking.time", "19:00", false)
	r.Set(ScopeTurn, "booking.time", "20:00", false) // latest wins

	res := r.Resolve(nil)
	if res.Available["merchant.name"] != "Cafe XYZ" {
		t.Errorf("expected workflow overrides tenant, got %v", res.Available["merchant.name"])
	}
	if res.Available["booking.time"] != "20:00" {
		t.Errorf("expected turn overrides state, got %v", res.Available["booking.time"])
	}
}

func TestContextResolverMissing(t *testing.T) {
	r := NewContextResolver()
	r.Set(ScopeMemory, "customer.name", "Baim", false)
	r.Set(ScopeWorkflow, "booking.date", "2026-08-27", false)

	res := r.Resolve([]string{"customer.name", "customer.address", "booking.date"})
	// customer.name & booking.date present; customer.address missing
	if len(res.Missing) != 1 || res.Missing[0] != "customer.address" {
		t.Errorf("expected only customer.address missing, got %v", res.Missing)
	}
}

func TestContextResolverMemoryWorkflowSplit(t *testing.T) {
	r := NewContextResolver()
	r.Set(ScopeMemory, "customer.address", "Jl. Melati", false)
	r.Set(ScopeWorkflow, "order.id", "ORD-1", false)

	res := r.Resolve(nil)
	if res.Memory["customer.address"] != "Jl. Melati" {
		t.Errorf("memory not preserved: %v", res.Memory)
	}
	if res.WorkflowData["order.id"] != "ORD-1" {
		t.Errorf("workflow data not preserved: %v", res.WorkflowData)
	}
	// workflow data should NOT be in memory and vice versa
	if _, ok := res.Memory["order.id"]; ok {
		t.Error("workflow data leaked into memory")
	}
	if _, ok := res.WorkflowData["customer.address"]; ok {
		t.Error("memory leaked into workflow data")
	}
}

func TestContextResolverSensitive(t *testing.T) {
	r := NewContextResolver()
	r.Set(ScopeTurn, "payment.token", "tok_secret", true)
	e, ok := r.Entry(ScopeTurn, "payment.token")
	if !ok {
		t.Fatal("entry not found")
	}
	if !e.Sensitive {
		t.Error("expected sensitive flag true")
	}
}
