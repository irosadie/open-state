package services

import (
	"context"
	"testing"
	"time"

	"github.com/irosadie/open-state/api/internal/application/dtos"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	domain "github.com/irosadie/open-state/go-shared/domain"
)

type adminIdentityRepoFake struct {
	tenant    *entities.Tenant
	members   map[string]*entities.TenantMembership
	lastScope string
	commits   int
}

type adminAuditRepoFake struct {
	entries []repositories.AppendAuditLogInput
}

func (f *adminAuditRepoFake) Append(_ context.Context, _ string, input repositories.AppendAuditLogInput) (*entities.AuditLog, error) {
	f.entries = append(f.entries, input)
	return &entities.AuditLog{}, nil
}
func (f *adminAuditRepoFake) ListByTenant(context.Context, string) ([]entities.AuditLog, error) {
	return nil, nil
}
func (f *adminAuditRepoFake) ListByAction(context.Context, string, entities.AuditAction) ([]entities.AuditLog, error) {
	return nil, nil
}
func (f *adminAuditRepoFake) ListByResource(context.Context, string, string, string) ([]entities.AuditLog, error) {
	return nil, nil
}
func (f *adminAuditRepoFake) ListFiltered(context.Context, string, repositories.AuditFilter) ([]entities.AuditLog, error) {
	return nil, nil
}
func (f *adminAuditRepoFake) CountFiltered(context.Context, string, repositories.AuditFilter) (int64, error) {
	return 0, nil
}

func (f *adminIdentityRepoFake) FindTenantByID(_ context.Context, tenantID string) (*entities.Tenant, error) {
	if f.tenant == nil || f.tenant.ID != tenantID {
		return nil, domain.NewNotFound("tenant not found")
	}
	copy := *f.tenant
	return &copy, nil
}

func (f *adminIdentityRepoFake) UpdateTenantProfile(_ context.Context, tenantID, name, slug, description string) (*entities.Tenant, error) {
	if f.tenant == nil || f.tenant.ID != tenantID {
		return nil, domain.NewNotFound("tenant not found")
	}
	f.tenant.Name, f.tenant.Slug, f.tenant.Description = name, slug, description
	f.tenant.UpdatedAt = f.tenant.UpdatedAt.Add(time.Minute)
	return f.FindTenantByID(context.Background(), tenantID)
}

func (f *adminIdentityRepoFake) ListMemberships(_ context.Context, tenantID string, search *string, offset, limit int) ([]entities.TenantMembership, error) {
	f.lastScope = tenantID
	result := make([]entities.TenantMembership, 0, limit)
	for _, member := range f.members {
		if member.TenantID != tenantID || (search != nil && *search != member.Name && *search != member.Email) {
			continue
		}
		result = append(result, *member)
	}
	if offset >= len(result) {
		return []entities.TenantMembership{}, nil
	}
	end := offset + limit
	if end > len(result) {
		end = len(result)
	}
	return result[offset:end], nil
}

func (f *adminIdentityRepoFake) CountMemberships(_ context.Context, tenantID string, _ *string) (int64, error) {
	var count int64
	for _, member := range f.members {
		if member.TenantID == tenantID {
			count++
		}
	}
	return count, nil
}

func (f *adminIdentityRepoFake) FindMembership(_ context.Context, tenantID, userID string) (*entities.TenantMembership, error) {
	member, ok := f.members[userID]
	if !ok || member.TenantID != tenantID {
		return nil, domain.NewNotFound("membership not found")
	}
	copy := *member
	return &copy, nil
}

func (f *adminIdentityRepoFake) CountOwners(_ context.Context, tenantID string) (int64, error) {
	var count int64
	for _, member := range f.members {
		if member.TenantID == tenantID && member.Role == entities.UserRoleOwner {
			count++
		}
	}
	return count, nil
}

func (f *adminIdentityRepoFake) AssignMembershipRole(_ context.Context, tenantID, userID string, role entities.UserRole) (*entities.TenantMembership, error) {
	member, err := f.FindMembership(context.Background(), tenantID, userID)
	if err != nil {
		return nil, err
	}
	f.members[userID].Role = role
	member.Role = role
	return member, nil
}

func (f *adminIdentityRepoFake) RemoveMembership(_ context.Context, tenantID, userID string) error {
	member, err := f.FindMembership(context.Background(), tenantID, userID)
	if err != nil {
		return err
	}
	delete(f.members, member.UserID)
	return nil
}

func (f *adminIdentityRepoFake) WithTx(_ context.Context, fn func(repositories.IAdminRepository) error) error {
	if err := fn(f); err != nil {
		return err
	}
	f.commits++
	return nil
}

func newAdminIdentityRepoFake() *adminIdentityRepoFake {
	now := time.Date(2026, 8, 29, 10, 0, 0, 0, time.UTC)
	return &adminIdentityRepoFake{
		tenant: &entities.Tenant{ID: "tenant-a", Name: "Tenant A", Slug: "tenant-a", CreatedAt: now, UpdatedAt: now},
		members: map[string]*entities.TenantMembership{
			"owner":  {UserID: "owner", TenantID: "tenant-a", Role: entities.UserRoleOwner, Name: "Owner", Email: "owner@example.com", CreatedAt: now, UpdatedAt: now},
			"editor": {UserID: "editor", TenantID: "tenant-a", Role: entities.UserRoleEditor, Name: "Editor", Email: "editor@example.com", CreatedAt: now, UpdatedAt: now},
			"other":  {UserID: "other", TenantID: "tenant-b", Role: entities.UserRoleOwner, Name: "Other", Email: "other@example.com", CreatedAt: now, UpdatedAt: now},
		},
	}
}

func TestAdminIdentityServiceScopesMembershipsToTenant(t *testing.T) {
	repo := newAdminIdentityRepoFake()
	service := NewAdminIdentityService(repo, nil)

	page, err := service.ListMemberships(context.Background(), "tenant-a", "", 1, 20)
	if err != nil {
		t.Fatalf("list memberships: %v", err)
	}
	if page.Total != 2 || len(page.Data) != 2 {
		t.Fatalf("expected two tenant-a memberships, got total=%d data=%d", page.Total, len(page.Data))
	}
	if repo.lastScope != "tenant-a" {
		t.Fatalf("expected tenant scope tenant-a, got %q", repo.lastScope)
	}
}

func TestAdminIdentityServiceProtectsLastOwner(t *testing.T) {
	repo := newAdminIdentityRepoFake()
	service := NewAdminIdentityService(repo, nil)

	_, err := service.UpdateMembershipRole(context.Background(), "tenant-a", "actor", "owner", "VIEWER", nil)
	assertDomainCode(t, err, domain.ErrConflict)
	if repo.commits != 0 {
		t.Fatal("last-owner rejection must not commit")
	}

	if err := service.RemoveMembership(context.Background(), "tenant-a", "actor", "owner", nil); err == nil {
		t.Fatal("expected last-owner removal to fail")
	} else {
		assertDomainCode(t, err, domain.ErrConflict)
	}
}

func TestAdminIdentityServiceValidatesRoleAndTenantPatch(t *testing.T) {
	repo := newAdminIdentityRepoFake()
	service := NewAdminIdentityService(repo, nil)

	_, err := service.UpdateMembershipRole(context.Background(), "tenant-a", "actor", "editor", "SUPERUSER", nil)
	assertDomainCode(t, err, domain.ErrValidation)

	_, err = service.UpdateTenant(context.Background(), "tenant-a", "actor", dtos.UpdateTenantRequest{Slug: stringPtr("not valid")}, nil)
	assertDomainCode(t, err, domain.ErrValidation)
}

func TestAdminIdentityServiceAuditsAcceptedAndRejectedMutations(t *testing.T) {
	repo := newAdminIdentityRepoFake()
	auditRepo := &adminAuditRepoFake{}
	service := NewAdminIdentityService(repo, NewAuditWriter(auditRepo, nil, nil))

	correlationID := "corr-1"
	_, err := service.UpdateTenant(context.Background(), "tenant-a", "actor", dtos.UpdateTenantRequest{Name: stringPtr("Tenant Updated")}, &correlationID)
	if err != nil {
		t.Fatalf("update tenant: %v", err)
	}
	_, err = service.UpdateMembershipRole(context.Background(), "tenant-a", "actor", "editor", "invalid", &correlationID)
	assertDomainCode(t, err, domain.ErrValidation)

	if len(auditRepo.entries) != 2 {
		t.Fatalf("expected accepted and rejected audit entries, got %d", len(auditRepo.entries))
	}
	if auditRepo.entries[0].Action != entities.AuditActionTenantUpdated || auditRepo.entries[0].ResourceType != "tenant" {
		t.Fatalf("unexpected tenant audit entry: %#v", auditRepo.entries[0])
	}
	if auditRepo.entries[1].Action != entities.AuditActionRoleUpdated || auditRepo.entries[1].ResourceType != "role_assignment" {
		t.Fatalf("unexpected role audit entry: %#v", auditRepo.entries[1])
	}
}

func assertDomainCode(t *testing.T, err error, code string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected %s error", code)
	}
	value, ok := err.(*domain.DomainError)
	if !ok || value.Code != code {
		t.Fatalf("expected domain code %s, got %T %v", code, err, err)
	}
}
