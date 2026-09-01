package services

import (
	"context"
	"strings"

	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	domain "github.com/irosadie/open-state/go-shared/domain"
)

// IntentService resolves conversation intents to their workflows (PRD 38, 171).
// It queries the published workflow definitions within a tenant+project, so
// `resolve_intent` returns real workflow data instead of a hardcoded stub.
type IntentService struct {
	intents   repositories.IIntentRepository
	workflows repositories.IWorkflowRepository
}

// NewIntentService builds an IntentService.
func NewIntentService(intents repositories.IIntentRepository, workflows repositories.IWorkflowRepository) *IntentService {
	return &IntentService{intents: intents, workflows: workflows}
}

// ListIntents returns the published workflow mappings in a tenant+project.
func (s *IntentService) ListIntents(ctx context.Context, tenantID, projectID string) ([]entities.Intent, error) {
	if strings.TrimSpace(tenantID) == "" || strings.TrimSpace(projectID) == "" {
		return nil, domain.NewValidation("tenant and project are required")
	}
	return s.intents.ListRoutable(ctx, tenantID, projectID)
}

// ResolveIntent returns the workflow mapped to a canonical intent key within a
// tenant+project, or a not-found error. Workflow IDs and slugs are not accepted.
func (s *IntentService) ResolveIntent(ctx context.Context, tenantID, projectID, intentID string) (*entities.Workflow, error) {
	if strings.TrimSpace(tenantID) == "" || strings.TrimSpace(projectID) == "" {
		return nil, domain.NewValidation("tenant and project are required")
	}
	key := strings.ToUpper(strings.TrimSpace(intentID))
	if key == "" {
		return nil, domain.NewValidation("intent is required")
	}
	intent, err := s.intents.FindRoutable(ctx, tenantID, projectID, key)
	if err != nil {
		return nil, err
	}
	wf, err := s.workflows.FindByID(ctx, tenantID, projectID, intent.WorkflowID)
	if err != nil {
		return nil, err
	}
	return wf, nil
}
