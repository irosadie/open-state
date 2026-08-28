package services

import (
	"context"

	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	domain "github.com/irosadie/open-state/go-shared/domain"
)

// IntentService resolves conversation intents to their workflows (PRD 38, 171).
// It queries the published workflow definitions within a tenant+project, so
// `resolve_intent` returns real workflow data instead of a hardcoded stub.
type IntentService struct {
	workflows repositories.IWorkflowRepository
}

// NewIntentService builds an IntentService.
func NewIntentService(workflows repositories.IWorkflowRepository) *IntentService {
	return &IntentService{workflows: workflows}
}

// ListIntents returns the workflows in a tenant+project as candidate intents.
func (s *IntentService) ListIntents(ctx context.Context, tenantID, projectID string) ([]entities.Workflow, error) {
	return s.workflows.ListByTenant(ctx, tenantID, projectID)
}

// ResolveIntent returns the workflow whose id (or slug) matches the intent within
// a tenant+project, or a not-found error.
func (s *IntentService) ResolveIntent(ctx context.Context, tenantID, projectID, intentID string) (*entities.Workflow, error) {
	if intentID == "" {
		return nil, domain.NewValidation("intent is required")
	}
	wf, err := s.workflows.FindByID(ctx, tenantID, projectID, intentID)
	if err == nil {
		return wf, nil
	}
	// Fall back to slug resolution.
	wf, err2 := s.workflows.FindBySlug(ctx, tenantID, projectID, intentID)
	if err2 != nil {
		return nil, domain.NewNotFound("intent not found")
	}
	return wf, nil
}
