package services

import (
	"context"
	"testing"

	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	domain "github.com/irosadie/open-state/go-shared/domain"
)

type fakeIntentRepository struct {
	intents []entities.Intent
}

func (f *fakeIntentRepository) ListRoutable(_ context.Context, tenantID, projectID string) ([]entities.Intent, error) {
	out := make([]entities.Intent, 0)
	for _, intent := range f.intents {
		if intent.TenantID == tenantID && intent.ProjectID == projectID {
			out = append(out, intent)
		}
	}
	return out, nil
}

func (f *fakeIntentRepository) FindRoutable(_ context.Context, tenantID, projectID, key string) (*entities.Intent, error) {
	for _, intent := range f.intents {
		if intent.TenantID == tenantID && intent.ProjectID == projectID && intent.Key == key {
			copy := intent
			return &copy, nil
		}
	}
	return nil, domain.NewNotFound("intent not found")
}

func (f *fakeIntentRepository) Upsert(_ context.Context, tenantID, projectID, workflowID, key, name, description string, examples []string) (*entities.Intent, error) {
	intent := entities.Intent{
		TenantID: tenantID, ProjectID: projectID, WorkflowID: workflowID,
		Key: key, Name: name, Description: description, Examples: examples,
	}
	f.intents = append(f.intents, intent)
	return &intent, nil
}

var _ repositories.IIntentRepository = (*fakeIntentRepository)(nil)

func TestIntentServiceListIntents(t *testing.T) {
	repo := &fakeIntentRepository{intents: []entities.Intent{
		{TenantID: "tenant-1", ProjectID: "project-padel", Key: "BOOKING_PADEL", WorkflowSlug: "padel-court-booking"},
		{TenantID: "tenant-1", ProjectID: "project-food", Key: "ORDER_FOOD", WorkflowSlug: "order-food"},
	}}
	svc := NewIntentService(repo, &fakeWorkflowRepo{})

	intents, err := svc.ListIntents(context.Background(), "tenant-1", "project-padel")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(intents) != 1 || intents[0].Key != "BOOKING_PADEL" {
		t.Fatalf("unexpected intents: %+v", intents)
	}

	if _, err := svc.ListIntents(context.Background(), "tenant-1", ""); err == nil {
		t.Fatal("expected scope validation error")
	}
}

func TestIntentServiceResolvesCanonicalKeyOnly(t *testing.T) {
	repo := &fakeIntentRepository{intents: []entities.Intent{{
		TenantID: "tenant-1", ProjectID: "project-padel", Key: "BOOKING_PADEL", WorkflowID: "wf-padel",
	}}}
	wfRepo := &fakeWorkflowRepo{workflows: map[string]*entities.Workflow{
		"wf-padel": {ID: "wf-padel", TenantID: "tenant-1", ProjectID: "project-padel", Slug: "padel-court-booking", Status: entities.WorkflowPublished},
	}}
	svc := NewIntentService(repo, wfRepo)

	wf, err := svc.ResolveIntent(context.Background(), "tenant-1", "project-padel", " booking_padel ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if wf.ID != "wf-padel" {
		t.Fatalf("unexpected workflow: %+v", wf)
	}

	if _, err := svc.ResolveIntent(context.Background(), "tenant-1", "project-padel", "padel-court-booking"); err == nil {
		t.Fatal("expected workflow slug to be rejected")
	}
}
