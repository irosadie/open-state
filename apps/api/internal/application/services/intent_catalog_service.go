package services

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/irosadie/open-state/api/internal/application/dtos"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	domain "github.com/irosadie/open-state/go-shared/domain"
)

const intentCatalogDefaultProjectSlug = "default"

// IntentCatalogService adapts the persisted intent catalog for authenticated
// HTTP clients. It resolves project scope without creating records and delegates
// routability filtering to the existing IntentService/repository path.
type IntentCatalogService struct {
	intents  *IntentService
	projects repositories.IProjectRepository
}

// NewIntentCatalogService builds the read-only HTTP intent catalog service.
func NewIntentCatalogService(intents *IntentService, projects repositories.IProjectRepository) *IntentCatalogService {
	return &IntentCatalogService{intents: intents, projects: projects}
}

// List returns published intent mappings for an explicit project or the
// tenant's existing default project. A read does not provision a project.
func (s *IntentCatalogService) List(ctx context.Context, tenantID, projectID string) (*dtos.IntentListDTO, error) {
	tenantID = strings.TrimSpace(tenantID)
	projectID = strings.TrimSpace(projectID)
	if tenantID == "" {
		return nil, domain.NewValidation("tenant is required")
	}

	project, err := s.resolveProject(ctx, tenantID, projectID)
	if err != nil {
		return nil, err
	}
	intents, err := s.intents.ListIntents(ctx, tenantID, project.ID)
	if err != nil {
		return nil, err
	}

	data := make([]dtos.IntentDTO, 0, len(intents))
	for i := range intents {
		data = append(data, toIntentDTO(&intents[i]))
	}
	return &dtos.IntentListDTO{Data: data}, nil
}

func (s *IntentCatalogService) resolveProject(ctx context.Context, tenantID, projectID string) (*entities.Project, error) {
	if projectID != "" {
		if _, err := uuid.Parse(projectID); err != nil {
			return nil, domain.NewValidation("projectId must be a valid UUID")
		}
		return s.projects.FindByID(ctx, tenantID, projectID)
	}
	return s.projects.FindBySlug(ctx, tenantID, intentCatalogDefaultProjectSlug)
}

func toIntentDTO(intent *entities.Intent) dtos.IntentDTO {
	examples := intent.Examples
	if examples == nil {
		examples = []string{}
	}
	return dtos.IntentDTO{
		ID:           intent.ID,
		TenantID:     intent.TenantID,
		ProjectID:    intent.ProjectID,
		WorkflowID:   intent.WorkflowID,
		Key:          intent.Key,
		Name:         intent.Name,
		Description:  intent.Description,
		Examples:     examples,
		WorkflowSlug: intent.WorkflowSlug,
	}
}
