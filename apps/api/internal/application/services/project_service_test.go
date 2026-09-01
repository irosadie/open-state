package services

import (
	"context"
	"testing"
	"time"

	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
)

type projectListRepository struct {
	projects []entities.Project
	tenantID string
}

func (r *projectListRepository) Create(context.Context, string, string, string, entities.ProjectStatus) (*entities.Project, error) {
	return nil, nil
}

func (r *projectListRepository) FindByID(context.Context, string, string) (*entities.Project, error) {
	return nil, nil
}

func (r *projectListRepository) FindBySlug(context.Context, string, string) (*entities.Project, error) {
	return nil, nil
}

func (r *projectListRepository) ListByTenant(_ context.Context, tenantID string) ([]entities.Project, error) {
	r.tenantID = tenantID
	return r.projects, nil
}

var _ repositories.IProjectRepository = (*projectListRepository)(nil)

func TestProjectServiceListsTenantProjects(t *testing.T) {
	createdAt := time.Date(2026, time.August, 31, 12, 0, 0, 0, time.UTC)
	repo := &projectListRepository{projects: []entities.Project{
		{ID: "project-1", TenantID: "tenant-1", Name: "Padel", Slug: "padel", Status: entities.ProjectActive, CreatedAt: createdAt, UpdatedAt: createdAt},
	}}

	result, err := NewProjectService(repo).List(context.Background(), " tenant-1 ")
	if err != nil {
		t.Fatalf("list projects: %v", err)
	}
	if repo.tenantID != "tenant-1" {
		t.Fatalf("expected trimmed tenant ID, got %q", repo.tenantID)
	}
	if len(result.Data) != 1 || result.Data[0].Name != "Padel" || result.Data[0].CreatedAt != "2026-08-31T12:00:00Z" {
		t.Fatalf("unexpected project response: %+v", result.Data)
	}
}

func TestProjectServiceRequiresTenant(t *testing.T) {
	_, err := NewProjectService(&projectListRepository{}).List(context.Background(), " ")
	if err == nil {
		t.Fatal("expected missing tenant validation")
	}
}
