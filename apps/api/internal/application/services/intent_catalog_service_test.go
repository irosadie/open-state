package services

import (
	"context"
	"errors"
	"testing"

	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	domain "github.com/irosadie/open-state/go-shared/domain"
)

type fakeIntentProjectRepository struct {
	projects    []entities.Project
	createCalls int
}

func (f *fakeIntentProjectRepository) Create(_ context.Context, tenantID, name, slug string, status entities.ProjectStatus) (*entities.Project, error) {
	f.createCalls++
	project := entities.Project{ID: "created-project", TenantID: tenantID, Name: name, Slug: slug, Status: status}
	f.projects = append(f.projects, project)
	return &project, nil
}

func (f *fakeIntentProjectRepository) FindByID(_ context.Context, tenantID, id string) (*entities.Project, error) {
	for _, project := range f.projects {
		if project.TenantID == tenantID && project.ID == id {
			copy := project
			return &copy, nil
		}
	}
	return nil, domain.NewNotFound("project not found")
}

func (f *fakeIntentProjectRepository) FindBySlug(_ context.Context, tenantID, slug string) (*entities.Project, error) {
	for _, project := range f.projects {
		if project.TenantID == tenantID && project.Slug == slug {
			copy := project
			return &copy, nil
		}
	}
	return nil, domain.NewNotFound("project not found")
}

func (f *fakeIntentProjectRepository) ListByTenant(_ context.Context, tenantID string) ([]entities.Project, error) {
	projects := make([]entities.Project, 0)
	for _, project := range f.projects {
		if project.TenantID == tenantID {
			projects = append(projects, project)
		}
	}
	return projects, nil
}

var _ repositories.IProjectRepository = (*fakeIntentProjectRepository)(nil)

func TestIntentCatalogServiceListsDefaultProjectWithoutProvisioning(t *testing.T) {
	const tenantID = "00000000-0000-0000-0000-000000000001"
	const projectID = "00000000-0000-0000-0000-000000000002"
	projects := &fakeIntentProjectRepository{projects: []entities.Project{{
		ID: projectID, TenantID: tenantID, Slug: "default", Status: entities.ProjectActive,
	}}}
	intentRepo := &fakeIntentRepository{intents: []entities.Intent{{
		ID: "intent-1", TenantID: tenantID, ProjectID: projectID, WorkflowID: "workflow-1",
		Key: "BOOKING_PADEL", Name: "Booking Padel", Examples: []string{"saya mau order lapangan"}, WorkflowSlug: "padel-court-booking",
	}}}

	svc := NewIntentCatalogService(NewIntentService(intentRepo, &fakeWorkflowRepo{}), projects)
	result, err := svc.List(context.Background(), tenantID, "")
	if err != nil {
		t.Fatalf("list default intents: %v", err)
	}
	if projects.createCalls != 0 {
		t.Fatalf("expected read path not to create a project, got %d calls", projects.createCalls)
	}
	if len(result.Data) != 1 || result.Data[0].Key != "BOOKING_PADEL" {
		t.Fatalf("unexpected catalog: %+v", result.Data)
	}
	if result.Data[0].WorkflowID != "workflow-1" || result.Data[0].WorkflowSlug != "padel-court-booking" {
		t.Fatalf("expected workflow mapping in DTO: %+v", result.Data[0])
	}
}

func TestIntentCatalogServiceUsesExplicitTenantProject(t *testing.T) {
	const tenantID = "00000000-0000-0000-0000-000000000001"
	const projectID = "00000000-0000-0000-0000-000000000003"
	projects := &fakeIntentProjectRepository{projects: []entities.Project{{
		ID: projectID, TenantID: tenantID, Slug: "padel", Status: entities.ProjectActive,
	}}}
	intentRepo := &fakeIntentRepository{intents: []entities.Intent{{
		TenantID: tenantID, ProjectID: projectID, Key: "BOOKING_PADEL",
	}}}

	svc := NewIntentCatalogService(NewIntentService(intentRepo, &fakeWorkflowRepo{}), projects)
	result, err := svc.List(context.Background(), tenantID, projectID)
	if err != nil {
		t.Fatalf("list explicit project intents: %v", err)
	}
	if len(result.Data) != 1 || result.Data[0].ProjectID != projectID {
		t.Fatalf("unexpected explicit project catalog: %+v", result.Data)
	}
}

func TestIntentCatalogServiceRejectsInvalidOrCrossTenantProject(t *testing.T) {
	const tenantID = "00000000-0000-0000-0000-000000000001"
	const otherProjectID = "00000000-0000-0000-0000-000000000004"
	projects := &fakeIntentProjectRepository{projects: []entities.Project{{
		ID: otherProjectID, TenantID: "00000000-0000-0000-0000-000000000099", Slug: "other", Status: entities.ProjectActive,
	}}}
	svc := NewIntentCatalogService(NewIntentService(&fakeIntentRepository{}, &fakeWorkflowRepo{}), projects)

	if _, err := svc.List(context.Background(), tenantID, "not-a-uuid"); err == nil {
		t.Fatal("expected invalid project id validation error")
	} else {
		var domainErr *domain.DomainError
		if !errors.As(err, &domainErr) || domainErr.Code != domain.ErrValidation {
			t.Fatalf("expected validation error, got %v", err)
		}
	}

	if _, err := svc.List(context.Background(), tenantID, otherProjectID); err == nil {
		t.Fatal("expected cross-tenant project lookup to fail")
	} else {
		var domainErr *domain.DomainError
		if !errors.As(err, &domainErr) || domainErr.Code != domain.ErrNotFound {
			t.Fatalf("expected not found error, got %v", err)
		}
	}
}

func TestIntentCatalogServiceReturnsEmptyCatalog(t *testing.T) {
	const tenantID = "00000000-0000-0000-0000-000000000001"
	const projectID = "00000000-0000-0000-0000-000000000005"
	projects := &fakeIntentProjectRepository{projects: []entities.Project{{
		ID: projectID, TenantID: tenantID, Slug: "empty", Status: entities.ProjectActive,
	}}}
	svc := NewIntentCatalogService(NewIntentService(&fakeIntentRepository{}, &fakeWorkflowRepo{}), projects)

	result, err := svc.List(context.Background(), tenantID, projectID)
	if err != nil {
		t.Fatalf("list empty catalog: %v", err)
	}
	if result == nil || result.Data == nil || len(result.Data) != 0 {
		t.Fatalf("expected non-nil empty data array, got %+v", result)
	}
}
