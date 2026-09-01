package capability

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"

	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/go-shared/domain"
)

// ---------------------------------------------------------------------------
// In-memory fake for ICapabilityRepository (DB-agnostic proof).
// ---------------------------------------------------------------------------

type fakeCapabilityRepo struct {
	byName   map[string]*entities.Capability
	bindings []entities.CapabilityBinding
}

func (f *fakeCapabilityRepo) Create(_ context.Context, _ string, name string, _ *string, pt entities.ProviderType, pid, _ *string, is, os []byte, _ int, _ *string) (*entities.Capability, error) {
	if f.byName == nil {
		f.byName = map[string]*entities.Capability{}
	}
	cap := &entities.Capability{Name: name, ProviderType: pt, InputSchema: is, OutputSchema: os, Status: entities.CapabilityActive}
	if pid != nil {
		cap.ProviderID = sql.NullString{String: *pid, Valid: true}
	}
	f.byName[name] = cap
	return cap, nil
}
func (f *fakeCapabilityRepo) FindByID(_ context.Context, _, id string) (*entities.Capability, error) {
	for _, c := range f.byName {
		if c.ID == id {
			return c, nil
		}
	}
	return nil, domain.NewNotFound("capability not found")
}
func (f *fakeCapabilityRepo) FindByName(_ context.Context, _, name string) (*entities.Capability, error) {
	c, ok := f.byName[name]
	if !ok {
		return nil, domain.NewNotFound("capability not found")
	}
	return c, nil
}
func (f *fakeCapabilityRepo) ListByTenant(_ context.Context, _ string) ([]entities.Capability, error) {
	out := []entities.Capability{}
	for _, c := range f.byName {
		out = append(out, *c)
	}
	return out, nil
}
func (f *fakeCapabilityRepo) ListByTenantFiltered(_ context.Context, _ string, _ entities.ProviderType, _ entities.CapabilityStatus) ([]entities.Capability, error) {
	return f.ListByTenant(context.Background(), "")
}
func (f *fakeCapabilityRepo) Update(_ context.Context, _ string, _ string, _ *string, _ entities.ProviderType, _, _ *string, _, _ []byte, _ entities.CapabilityStatus, _ int, _ *string) (*entities.Capability, error) {
	return nil, nil
}
func (f *fakeCapabilityRepo) UpdateStatus(_ context.Context, _, _ string, _ entities.CapabilityStatus) (*entities.Capability, error) {
	return nil, nil
}
func (f *fakeCapabilityRepo) Disable(_ context.Context, _, _ string) (*entities.Capability, error) {
	return nil, nil
}
func (f *fakeCapabilityRepo) Bind(_ context.Context, _, _ string, st entities.BindingScopeType, _ string, perm entities.BindingPermission) (*entities.CapabilityBinding, error) {
	f.bindings = append(f.bindings, entities.CapabilityBinding{ScopeType: st, Permission: perm})
	return nil, nil
}
func (f *fakeCapabilityRepo) ListBindingsByCapability(_ context.Context, _, _ string) ([]entities.CapabilityBinding, error) {
	return f.bindings, nil
}
func (f *fakeCapabilityRepo) ListBindingsByScope(_ context.Context, _ string, _ entities.BindingScopeType, _ string) ([]entities.CapabilityBinding, error) {
	return f.bindings, nil
}
func (f *fakeCapabilityRepo) Unbind(_ context.Context, _, _ string) error { return nil }
func (f *fakeCapabilityRepo) UpsertPolicy(_ context.Context, _ string, _ entities.PolicyScopeType, _, _ string, _ []byte) (*entities.Policy, error) {
	return nil, nil
}
func (f *fakeCapabilityRepo) FindPolicyByType(_ context.Context, _ string, _ entities.PolicyScopeType, _, _ string) (*entities.Policy, error) {
	return nil, nil
}
func (f *fakeCapabilityRepo) ListPoliciesByScope(_ context.Context, _ string, _ entities.PolicyScopeType, _ string) ([]entities.Policy, error) {
	return nil, nil
}

// ---------------------------------------------------------------------------
// Resolver tests
// ---------------------------------------------------------------------------

func TestResolverResolveAllowed(t *testing.T) {
	pid := "mcp-payment"
	schema := []byte(`{"type":"object"}`)
	repo := &fakeCapabilityRepo{}
	_, _ = repo.Create(context.Background(), "t", "payment.create", nil, entities.ProviderTypeMCP, &pid, nil, schema, nil, 1, nil)
	r := NewCapabilityResolver(repo)

	res, err := r.Resolve(context.Background(), "t", "payment.create", "wf", "st")
	if err != nil {
		t.Fatalf("resolve error: %v", err)
	}
	if res.Name != "payment.create" || res.ProviderType != entities.ProviderTypeMCP {
		t.Errorf("unexpected resolution: %+v", res)
	}
}

func TestResolverDeniedMostRestrictive(t *testing.T) {
	pid := "mcp-payment"
	repo := &fakeCapabilityRepo{}
	_, _ = repo.Create(context.Background(), "t", "payment.create", nil, entities.ProviderTypeMCP, &pid, nil, nil, nil, 1, nil)
	// tenant allows, state denies → deny wins (state more specific)
	repo.bindings = []entities.CapabilityBinding{
		{ScopeType: entities.BindingScopeTenant, ScopeID: "t", Permission: entities.BindingPermissionAllow},
		{ScopeType: entities.BindingScopeState, ScopeID: "st", Permission: entities.BindingPermissionDeny},
	}
	r := NewCapabilityResolver(repo)

	_, err := r.Resolve(context.Background(), "t", "payment.create", "wf", "st")
	if err == nil {
		t.Fatal("expected denial error")
	}
	var ce *CapabilityError
	if !errors.As(err, &ce) || ce.Kind != ErrorKindUnauthorized {
		t.Errorf("expected unauthorized, got %v", err)
	}
}

func TestResolverUnknownCapability(t *testing.T) {
	r := NewCapabilityResolver(&fakeCapabilityRepo{})
	_, err := r.Resolve(context.Background(), "t", "does.not.exist", "wf", "st")
	if err == nil {
		t.Fatal("expected not-found → unauthorized")
	}
	var ce *CapabilityError
	if !errors.As(err, &ce) || ce.Code != "capability.unauthorized" {
		t.Errorf("expected unauthorized code, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// Idempotency store tests
// ---------------------------------------------------------------------------

func TestInMemoryIdempotency(t *testing.T) {
	s := NewInMemoryIdempotencyStore()
	ctx := context.Background()
	_, hit, _ := s.Get(ctx, "k1")
	if hit {
		t.Fatal("expected miss on empty store")
	}
	_ = s.Put(ctx, "k1", InvocationResult{Data: map[string]any{"ok": true}})
	res, hit, _ := s.Get(ctx, "k1")
	if !hit || res.Data["ok"] != true {
		t.Fatal("expected hit with stored result")
	}
}

func TestBuildIdempotencyKey(t *testing.T) {
	if k := buildIdempotencyKey("", "a1"); k != "" {
		t.Errorf("expected empty key, got %q", k)
	}
	if k := buildIdempotencyKey("wf", "a1"); k != "wf:a1" {
		t.Errorf("expected wf:a1, got %q", k)
	}
}

// ---------------------------------------------------------------------------
// Error classification tests
// ---------------------------------------------------------------------------

func TestCapabilityErrorRetryable(t *testing.T) {
	timeout := NewCapabilityError(ErrorKindTimeout, "capability.timeout", "x")
	if !timeout.Retryable() {
		t.Error("timeout should be retryable")
	}
	auth := NewCapabilityError(ErrorKindUnauthorized, "capability.unauthorized", "x")
	if auth.Retryable() {
		t.Error("unauthorized should not be retryable")
	}
}

func TestCodeForCapabilityEvent(t *testing.T) {
	cases := map[ErrorKind]string{
		ErrorKindTimeout: "capability.timeout", ErrorKindUnauthorized: "capability.unauthorized",
		ErrorKindValidation: "capability.validation_failed", ErrorKindUnavailable: "capability.unavailable",
		ErrorKindBusiness: "capability.business_error",
	}
	for k, want := range cases {
		if got := CodeForCapabilityEvent(k); got != want {
			t.Errorf("%s: got %s want %s", k, got, want)
		}
	}
}

// ---------------------------------------------------------------------------
// JSON helper (validators)
// ---------------------------------------------------------------------------

var _ = json.Valid
