package services

import (
	"context"
	"database/sql"
	"testing"

	"github.com/irosadie/open-state/api/internal/application/dtos"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	domain "github.com/irosadie/open-state/go-shared/domain"
)

// fakeWorkflowRepo is a minimal in-memory IWorkflowRepository for service tests.
type fakeWorkflowRepo struct {
	workflows map[string]*entities.Workflow
	versions  map[string][]*entities.WorkflowVersion
}

func (f *fakeWorkflowRepo) Create(_ context.Context, tenantID, projectID, slug, name string, description *string) (*entities.Workflow, error) {
	if f.workflows == nil {
		f.workflows = map[string]*entities.Workflow{}
	}
	for _, w := range f.workflows {
		if w.TenantID == tenantID && w.ProjectID == projectID && w.Slug == slug {
			return nil, domain.NewConflict("create workflow: already exists")
		}
	}
	w := &entities.Workflow{
		ID: "wf-1", TenantID: tenantID, ProjectID: projectID, Slug: slug, Name: name,
		Status: entities.WorkflowDraft, Version: 0,
	}
	if description != nil {
		w.Description = sql.NullString{String: *description, Valid: true}
	}
	f.workflows[w.ID] = w
	return w, nil
}

func (f *fakeWorkflowRepo) FindByID(_ context.Context, _, _, id string) (*entities.Workflow, error) {
	w, ok := f.workflows[id]
	if !ok {
		return nil, domain.NewNotFound("workflow not found")
	}
	return w, nil
}

func (f *fakeWorkflowRepo) FindBySlug(_ context.Context, _, _, _ string) (*entities.Workflow, error) {
	return nil, domain.NewNotFound("workflow not found")
}

func (f *fakeWorkflowRepo) ListByTenant(_ context.Context, tenantID, projectID string) ([]entities.Workflow, error) {
	out := []entities.Workflow{}
	for _, w := range f.workflows {
		if w.TenantID == tenantID && w.ProjectID == projectID {
			out = append(out, *w)
		}
	}
	return out, nil
}

func (f *fakeWorkflowRepo) UpdateStatus(_ context.Context, _, _, id string, _ entities.WorkflowStatus, expectedVersion int) (*entities.Workflow, error) {
	w, ok := f.workflows[id]
	if !ok {
		return nil, domain.NewNotFound("workflow not found")
	}
	if w.Version != expectedVersion {
		return nil, domain.NewConflict("optimistic lock conflict: resource changed")
	}
	w.Version++
	return w, nil
}

func (f *fakeWorkflowRepo) CreateVersion(_ context.Context, _, _, _ string, versionNo int, definition []byte, status entities.VersionStatus, isCurrent bool) (*entities.WorkflowVersion, error) {
	if f.versions == nil {
		f.versions = map[string][]*entities.WorkflowVersion{}
	}
	v := &entities.WorkflowVersion{
		ID: "ver-1", WorkflowID: "wf-1", VersionNo: versionNo, Definition: definition, Status: status, IsCurrent: isCurrent,
	}
	f.versions["wf-1"] = append(f.versions["wf-1"], v)
	return v, nil
}

func (f *fakeWorkflowRepo) Publish(_ context.Context, _, _, _ string, versionNo int, definition []byte, status entities.VersionStatus, expectedVersion int) (*entities.WorkflowVersion, error) {
	w := f.workflows["wf-1"]
	if w == nil {
		return nil, domain.NewNotFound("workflow not found")
	}
	if w.Version != expectedVersion {
		return nil, domain.NewConflict("optimistic lock conflict: resource changed")
	}
	if f.versions == nil {
		f.versions = map[string][]*entities.WorkflowVersion{}
	}
	v := &entities.WorkflowVersion{
		ID: "ver-1", WorkflowID: w.ID, VersionNo: versionNo, Definition: definition, Status: status, IsCurrent: true,
	}
	f.versions[w.ID] = append(f.versions[w.ID], v)
	w.CurrentVersion = versionNo
	w.Version++
	return v, nil
}

func (f *fakeWorkflowRepo) FindCurrentVersion(_ context.Context, _, _, _ string) (*entities.WorkflowVersion, error) {
	return nil, domain.NewNotFound("workflow version not found")
}

func (f *fakeWorkflowRepo) ListVersions(_ context.Context, _, _, workflowID string) ([]entities.WorkflowVersion, error) {
	vs := f.versions[workflowID]
	out := make([]entities.WorkflowVersion, 0, len(vs))
	for _, v := range vs {
		out = append(out, *v)
	}
	return out, nil
}

func (f *fakeWorkflowRepo) FindVersionByNumber(_ context.Context, _, _, _ string, _ int) (*entities.WorkflowVersion, error) {
	return nil, domain.NewNotFound("workflow version not found")
}

func (f *fakeWorkflowRepo) CreateState(_ context.Context, _, _, _ string, _ string, _ entities.StateKind, _ string, _, _ *string, _, _, _, _ []byte, _ bool) (*entities.State, error) {
	return nil, nil
}

func (f *fakeWorkflowRepo) ListStatesByVersion(_ context.Context, _, _, _ string) ([]entities.State, error) {
	return nil, nil
}

func (f *fakeWorkflowRepo) CreateTransition(_ context.Context, _, _, _ string, _ string, _, _ string, _ string, _ int, _ bool) (*entities.Transition, error) {
	return nil, nil
}

func (f *fakeWorkflowRepo) ListTransitionsByVersion(_ context.Context, _, _, _ string) ([]entities.Transition, error) {
	return nil, nil
}

func (f *fakeWorkflowRepo) CreateTransitionGuard(_ context.Context, _, _, _ string, _ string, _ string, _ []byte) (*entities.TransitionGuard, error) {
	return nil, nil
}

func (f *fakeWorkflowRepo) ListGuardsByTransition(_ context.Context, _, _, _ string) ([]entities.TransitionGuard, error) {
	return nil, nil
}

// fakeProjectRepo is a minimal in-memory IProjectRepository.
type fakeProjectRepo struct {
	projects map[string]*entities.Project
}

func (f *fakeProjectRepo) Create(_ context.Context, tenantID, name, slug string, status entities.ProjectStatus) (*entities.Project, error) {
	if f.projects == nil {
		f.projects = map[string]*entities.Project{}
	}
	p := &entities.Project{ID: "proj-1", TenantID: tenantID, Name: name, Slug: slug, Status: status}
	f.projects[p.Slug] = p
	return p, nil
}

func (f *fakeProjectRepo) FindByID(_ context.Context, _, _ string) (*entities.Project, error) {
	return nil, domain.NewNotFound("project not found")
}

func (f *fakeProjectRepo) FindBySlug(_ context.Context, _, slug string) (*entities.Project, error) {
	p, ok := f.projects[slug]
	if !ok {
		return nil, domain.NewNotFound("project not found")
	}
	return p, nil
}

func (f *fakeProjectRepo) ListByTenant(_ context.Context, _ string) ([]entities.Project, error) {
	return nil, nil
}

func newBuilderService() (*BuilderService, *fakeWorkflowRepo, *fakeProjectRepo) {
	wfRepo := &fakeWorkflowRepo{}
	projRepo := &fakeProjectRepo{}
	return NewBuilderService(wfRepo, projRepo, nil), wfRepo, projRepo
}

var _ repositories.IWorkflowRepository = (*fakeWorkflowRepo)(nil)
var _ repositories.IProjectRepository = (*fakeProjectRepo)(nil)

func TestCreateDraft(t *testing.T) {
	svc, _, _ := newBuilderService()
	ctx := context.Background()

	dto, err := svc.CreateDraft(ctx, "tenant-1", dtos.CreateWorkflowRequest{
		Slug: "padel", Name: "Padel Booking", Description: "booking flow",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dto.Status != string(entities.WorkflowDraft) {
		t.Fatalf("expected DRAFT, got %s", dto.Status)
	}
	if dto.ProjectID == "" {
		t.Fatal("expected default project to be resolved")
	}
}

func TestCreateDraftDuplicateSlug(t *testing.T) {
	svc, _, _ := newBuilderService()
	ctx := context.Background()

	if _, err := svc.CreateDraft(ctx, "tenant-1", dtos.CreateWorkflowRequest{Slug: "padel", Name: "A"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := svc.CreateDraft(ctx, "tenant-1", dtos.CreateWorkflowRequest{Slug: "padel", Name: "B"}); err == nil {
		t.Fatal("expected duplicate slug to fail")
	}
}

func TestCreateDraftRequiresSlugAndName(t *testing.T) {
	svc, _, _ := newBuilderService()
	ctx := context.Background()

	if _, err := svc.CreateDraft(ctx, "tenant-1", dtos.CreateWorkflowRequest{Name: "A"}); err == nil {
		t.Fatal("expected slug validation error")
	}
	if _, err := svc.CreateDraft(ctx, "tenant-1", dtos.CreateWorkflowRequest{Slug: "x"}); err == nil {
		t.Fatal("expected name validation error")
	}
}

func TestUpdateDraftOptimisticLock(t *testing.T) {
	svc, wfRepo, _ := newBuilderService()
	ctx := context.Background()

	created, err := svc.CreateDraft(ctx, "tenant-1", dtos.CreateWorkflowRequest{Slug: "padel", Name: "A"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = wfRepo // version begins at 0

	// stale version -> conflict
	if _, err := svc.UpdateDraft(ctx, "tenant-1", created.ProjectID, created.ID, dtos.UpdateWorkflowRequest{Version: 5}); err == nil {
		t.Fatal("expected optimistic lock conflict on stale version")
	}
	// current version -> success, version bumped
	updated, err := svc.UpdateDraft(ctx, "tenant-1", created.ProjectID, created.ID, dtos.UpdateWorkflowRequest{Version: created.Version})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Version != created.Version+1 {
		t.Fatalf("expected version to bump to %d, got %d", created.Version+1, updated.Version)
	}
}

func TestPublishCreatesImmutableVersion(t *testing.T) {
	svc, _, _ := newBuilderService()
	ctx := context.Background()

	created, err := svc.CreateDraft(ctx, "tenant-1", dtos.CreateWorkflowRequest{Slug: "padel", Name: "A"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	version, err := svc.Publish(ctx, "tenant-1", created.ProjectID, created.ID, "actor-1", dtos.PublishWorkflowRequest{
		Version:    created.Version,
		Definition: []byte(`{"nodes":[]}`),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if version.VersionNo != 1 {
		t.Fatalf("expected versionNo 1, got %d", version.VersionNo)
	}
	if !version.IsCurrent {
		t.Fatal("expected published version to be current")
	}
}

func TestPublishRequiresDefinition(t *testing.T) {
	svc, _, _ := newBuilderService()
	ctx := context.Background()

	created, err := svc.CreateDraft(ctx, "tenant-1", dtos.CreateWorkflowRequest{Slug: "padel", Name: "A"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := svc.Publish(ctx, "tenant-1", created.ProjectID, created.ID, "actor-1", dtos.PublishWorkflowRequest{Version: 0}); err == nil {
		t.Fatal("expected validation error for missing definition")
	}
}

func TestListVersions(t *testing.T) {
	svc, _, _ := newBuilderService()
	ctx := context.Background()

	created, err := svc.CreateDraft(ctx, "tenant-1", dtos.CreateWorkflowRequest{Slug: "padel", Name: "A"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := svc.Publish(ctx, "tenant-1", created.ProjectID, created.ID, "actor-1", dtos.PublishWorkflowRequest{Version: 0, Definition: []byte(`{}`)}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	versions, err := svc.ListVersions(ctx, "tenant-1", created.ProjectID, created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(versions) != 1 {
		t.Fatalf("expected 1 version, got %d", len(versions))
	}
}
