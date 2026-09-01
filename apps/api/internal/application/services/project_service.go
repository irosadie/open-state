package services

import (
	"context"
	"strings"
	"time"

	"github.com/irosadie/open-state/api/internal/application/dtos"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	domain "github.com/irosadie/open-state/go-shared/domain"
)

// ProjectService provides read-only tenant project discovery for clients that
// need to choose an existing project, including the State MCP API-key console.
type ProjectService struct {
	projects repositories.IProjectRepository
}

// NewProjectService builds a project discovery service.
func NewProjectService(projects repositories.IProjectRepository) *ProjectService {
	return &ProjectService{projects: projects}
}

// List returns every project owned by tenantID. The repository applies the
// tenant boundary; no project from another tenant can be returned.
func (s *ProjectService) List(ctx context.Context, tenantID string) (*dtos.ProjectListDTO, error) {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return nil, domain.NewValidation("tenant is required")
	}

	projects, err := s.projects.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	data := make([]dtos.ProjectDTO, 0, len(projects))
	for i := range projects {
		data = append(data, toProjectDTO(&projects[i]))
	}

	return &dtos.ProjectListDTO{Data: data}, nil
}

func toProjectDTO(project *entities.Project) dtos.ProjectDTO {
	return dtos.ProjectDTO{
		ID:        project.ID,
		TenantID:  project.TenantID,
		Name:      project.Name,
		Slug:      project.Slug,
		Status:    string(project.Status),
		CreatedAt: project.CreatedAt.Format(time.RFC3339),
		UpdatedAt: project.UpdatedAt.Format(time.RFC3339),
	}
}
