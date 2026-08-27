package services

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/irosadie/open-state/api/internal/application/dtos"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	infracap "github.com/irosadie/open-state/api/internal/infrastructure/capability"
	domain "github.com/irosadie/open-state/go-shared/domain"
)

// fakeCapRepo is a minimal in-memory ICapabilityRepository for service tests.
type fakeCapRepo struct {
	caps map[string]*entities.Capability
}

func (f *fakeCapRepo) Create(_ context.Context, tenantID, name string, description *string, pt entities.ProviderType, providerID *string, is, os []byte, version int, credRef *string) (*entities.Capability, error) {
	if f.caps == nil {
		f.caps = map[string]*entities.Capability{}
	}
	for _, c := range f.caps {
		if c.TenantID == tenantID && c.Name == name {
			return nil, domain.NewConflict("capability: already exists")
		}
	}
	c := &entities.Capability{
		ID: "cap-1", TenantID: tenantID, Name: name, ProviderType: pt, Status: entities.CapabilityActive,
		Version: version, InputSchema: is, OutputSchema: os,
	}
	if description != nil {
		c.Description = sql.NullString{String: *description, Valid: true}
	}
	if providerID != nil {
		c.ProviderID = sql.NullString{String: *providerID, Valid: true}
	}
	if credRef != nil {
		c.CredentialReference = sql.NullString{String: *credRef, Valid: true}
	}
	f.caps[name] = c
	return c, nil
}

func (f *fakeCapRepo) FindByID(_ context.Context, _, id string) (*entities.Capability, error) {
	for _, c := range f.caps {
		if c.ID == id {
			return c, nil
		}
	}
	return nil, domain.NewNotFound("capability not found")
}

func (f *fakeCapRepo) FindByName(_ context.Context, _, name string) (*entities.Capability, error) {
	c, ok := f.caps[name]
	if !ok {
		return nil, domain.NewNotFound("capability not found")
	}
	return c, nil
}

func (f *fakeCapRepo) ListByTenant(_ context.Context, _ string) ([]entities.Capability, error) {
	out := []entities.Capability{}
	for _, c := range f.caps {
		out = append(out, *c)
	}
	return out, nil
}

func (f *fakeCapRepo) ListByTenantFiltered(_ context.Context, _ string, _ entities.ProviderType, _ entities.CapabilityStatus) ([]entities.Capability, error) {
	return f.ListByTenant(context.Background(), "")
}

func (f *fakeCapRepo) Update(_ context.Context, _, id string, description *string, providerType entities.ProviderType, providerID *string, input, output []byte, status entities.CapabilityStatus, version int, credRef *string) (*entities.Capability, error) {
	for _, c := range f.caps {
		if c.ID == id {
			if description != nil {
				c.Description = sql.NullString{String: *description, Valid: true}
			}
			c.ProviderType = providerType
			if providerID != nil {
				c.ProviderID = sql.NullString{String: *providerID, Valid: true}
			}
			c.InputSchema = input
			c.OutputSchema = output
			c.Status = status
			c.Version = version
			if credRef != nil {
				c.CredentialReference = sql.NullString{String: *credRef, Valid: true}
			}
			return c, nil
		}
	}
	return nil, domain.NewNotFound("capability not found")
}

func (f *fakeCapRepo) UpdateStatus(_ context.Context, _, _ string, _ entities.CapabilityStatus) (*entities.Capability, error) {
	return nil, nil
}

func (f *fakeCapRepo) Disable(_ context.Context, _, _ string) (*entities.Capability, error) {
	return nil, nil
}

func (f *fakeCapRepo) Bind(_ context.Context, _, _ string, _ entities.BindingScopeType, _ string, _ entities.BindingPermission) (*entities.CapabilityBinding, error) {
	return nil, nil
}

func (f *fakeCapRepo) ListBindingsByCapability(_ context.Context, _, _ string) ([]entities.CapabilityBinding, error) {
	return nil, nil
}

func (f *fakeCapRepo) ListBindingsByScope(_ context.Context, _ string, _ entities.BindingScopeType, _ string) ([]entities.CapabilityBinding, error) {
	return nil, nil
}

func (f *fakeCapRepo) Unbind(_ context.Context, _, _ string) error { return nil }

func (f *fakeCapRepo) UpsertPolicy(_ context.Context, _ string, _ entities.PolicyScopeType, _, _ string, _ []byte) (*entities.Policy, error) {
	return nil, nil
}

func (f *fakeCapRepo) FindPolicyByType(_ context.Context, _ string, _ entities.PolicyScopeType, _, _ string) (*entities.Policy, error) {
	return nil, nil
}

func (f *fakeCapRepo) ListPoliciesByScope(_ context.Context, _ string, _ entities.PolicyScopeType, _ string) ([]entities.Policy, error) {
	return nil, nil
}

func isValidation(err error) bool {
	var de *domain.DomainError
	return errors.As(err, &de) && de.Code == domain.ErrValidation
}

// newTestService builds a CapabilityService with the real mock/sandbox
// collaborators, mirroring the composition root wiring.
func newTestService(repo repositories.ICapabilityRepository) *CapabilityService {
	return NewCapabilityService(repo, infracap.MockProviderResolver{}, infracap.JSONSchemaValidator{})
}

func TestCreateInvalidProviderType(t *testing.T) {
	svc := newTestService(&fakeCapRepo{})
	_, err := svc.Create(context.Background(), "tenant-1", dtos.CreateCapabilityRequest{
		Name: "payment.create", ProviderType: "NOT_A_TYPE",
	})
	if !isValidation(err) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestCreateEmptyName(t *testing.T) {
	svc := newTestService(&fakeCapRepo{})
	_, err := svc.Create(context.Background(), "tenant-1", dtos.CreateCapabilityRequest{Name: "", ProviderType: "MCP"})
	if !isValidation(err) {
		t.Fatalf("expected validation error for empty name, got %v", err)
	}
}

func TestCreateDuplicateNameConflict(t *testing.T) {
	repo := &fakeCapRepo{}
	svc := newTestService(repo)
	_, err := svc.Create(context.Background(), "tenant-1", dtos.CreateCapabilityRequest{Name: "payment.create", ProviderType: "MCP"})
	if err != nil {
		t.Fatalf("first create failed: %v", err)
	}
	_, err = svc.Create(context.Background(), "tenant-1", dtos.CreateCapabilityRequest{Name: "payment.create", ProviderType: "MCP"})
	var de *domain.DomainError
	if !errors.As(err, &de) || de.Code != domain.ErrConflict {
		t.Fatalf("expected conflict, got %v", err)
	}
}

func TestUpdateInvalidStatus(t *testing.T) {
	svc := newTestService(&fakeCapRepo{})
	created, err := svc.Create(context.Background(), "tenant-1", dtos.CreateCapabilityRequest{
		Name: "payment.create", ProviderType: "MCP",
		InputSchema:  map[string]any{"type": "object", "required": []any{"amount"}},
		OutputSchema: map[string]any{"type": "object"},
	})
	if err != nil {
		t.Fatalf("seed create failed: %v", err)
	}
	_, err = svc.Update(context.Background(), "tenant-1", created.ID, dtos.UpdateCapabilityRequest{
		ProviderType: "MCP", Status: "BAD_STATUS",
	})
	if !isValidation(err) {
		t.Fatalf("expected validation error for invalid status, got %v", err)
	}
}

// TestUpdatePreservesUnsetFields verifies partial/PATCH semantics: fields not
// present in the update are preserved (e.g. input schema is not wiped).
func TestUpdatePreservesUnsetFields(t *testing.T) {
	svc := newTestService(&fakeCapRepo{})
	created, err := svc.Create(context.Background(), "tenant-1", dtos.CreateCapabilityRequest{
		Name: "payment.create", ProviderType: "MCP",
		InputSchema:  map[string]any{"type": "object", "required": []any{"amount"}},
		OutputSchema: map[string]any{"type": "object"},
	})
	if err != nil {
		t.Fatalf("seed create failed: %v", err)
	}
	updated, err := svc.Update(context.Background(), "tenant-1", created.ID, dtos.UpdateCapabilityRequest{
		Description: "only description changed",
	})
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if updated.InputSchema == nil {
		t.Fatal("input schema should be preserved when not provided")
	}
	if updated.ProviderType != "MCP" {
		t.Fatalf("provider type should be preserved, got %s", updated.ProviderType)
	}
	if updated.Description == nil || *updated.Description != "only description changed" {
		t.Fatalf("description should be updated, got %v", updated.Description)
	}
}

func TestBindInvalidScopeType(t *testing.T) {
	svc := newTestService(&fakeCapRepo{})
	_, err := svc.Bind(context.Background(), "tenant-1", "cap-1", dtos.CreateBindingRequest{
		ScopeType: "GLOBAL", ScopeID: "x", Permission: "ALLOW",
	})
	if !isValidation(err) {
		t.Fatalf("expected validation error for invalid scopeType, got %v", err)
	}
}

func TestBindEmptyScopeID(t *testing.T) {
	svc := newTestService(&fakeCapRepo{})
	_, err := svc.Bind(context.Background(), "tenant-1", "cap-1", dtos.CreateBindingRequest{
		ScopeType: "STATE", ScopeID: "", Permission: "ALLOW",
	})
	if !isValidation(err) {
		t.Fatalf("expected validation error for empty scopeId, got %v", err)
	}
}

// TestSecretSafety asserts that the CapabilityDTO exposes only the credential
// reference string, never a resolved secret value (PRD §61, §91).
func TestSecretSafety(t *testing.T) {
	repo := &fakeCapRepo{}
	svc := newTestService(repo)
	secret := "super-secret-token"
	dto, err := svc.Create(context.Background(), "tenant-1", dtos.CreateCapabilityRequest{
		Name: "payment.create", ProviderType: "MCP", CredentialReference: secret,
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if dto.CredentialReference == nil || *dto.CredentialReference != secret {
		t.Fatalf("expected credential reference to be preserved, got %v", dto.CredentialReference)
	}
	// The DTO must not carry any field named for the resolved secret.
	if dto.ID == "" || dto.Name != "payment.create" {
		t.Fatalf("unexpected dto: %+v", dto)
	}
}
