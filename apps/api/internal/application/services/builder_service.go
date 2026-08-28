package services

import (
	"context"
	"errors"
	"time"

	"github.com/irosadie/open-state/api/internal/application/dtos"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	domain "github.com/irosadie/open-state/go-shared/domain"
)

// defaultProjectSlug is the fallback project created for a tenant when no
// projectId is supplied (PRD §3.1.1). It lets the State Builder persist drafts
// end-to-end before a project-management UI exists.
const defaultProjectSlug = "default"

// BuilderService orchestrates workflow-definition authoring operations exposed by
// the Builder API (PRD 146): create/get/update draft, list, publish, and
// list-versions. Every operation is tenant+project scoped (PRD §4, §96). It
// composes the workflow repository and, for default-project resolution, the
// project repository.
type BuilderService struct {
	workflows repositories.IWorkflowRepository
	projects  repositories.IProjectRepository
	audit     *AuditWriter
	now       func() time.Time
}

// NewBuilderService builds a BuilderService.
func NewBuilderService(workflows repositories.IWorkflowRepository, projects repositories.IProjectRepository, audit *AuditWriter) *BuilderService {
	return &BuilderService{workflows: workflows, projects: projects, audit: audit, now: time.Now}
}

// CreateDraft persists a new workflow definition draft for the tenant within a
// project (defaulted to the tenant's "default" project when absent).
func (s *BuilderService) CreateDraft(ctx context.Context, tenantID string, req dtos.CreateWorkflowRequest) (*dtos.WorkflowDTO, error) {
	if req.Slug == "" {
		return nil, domain.NewValidation("slug is required")
	}
	if req.Name == "" {
		return nil, domain.NewValidation("name is required")
	}

	projectID, err := s.resolveProject(ctx, tenantID, req.ProjectID)
	if err != nil {
		return nil, err
	}

	wf, err := s.workflows.Create(ctx, tenantID, projectID, req.Slug, req.Name, optional(req.Description))
	if err != nil {
		return nil, err
	}
	return toWorkflowDTO(wf), nil
}

// Get returns a single tenant/project-scoped workflow definition.
func (s *BuilderService) Get(ctx context.Context, tenantID, projectID, id string) (*dtos.WorkflowDTO, error) {
	pid, err := s.resolveProject(ctx, tenantID, projectID)
	if err != nil {
		return nil, err
	}
	wf, err := s.workflows.FindByID(ctx, tenantID, pid, id)
	if err != nil {
		return nil, err
	}
	return toWorkflowDTO(wf), nil
}

// List returns the tenant's workflow definitions within a project.
func (s *BuilderService) List(ctx context.Context, tenantID, projectID string) (*dtos.WorkflowListDTO, error) {
	pid, err := s.resolveProject(ctx, tenantID, projectID)
	if err != nil {
		return nil, err
	}
	list, err := s.workflows.ListByTenant(ctx, tenantID, pid)
	if err != nil {
		return nil, err
	}
	data := make([]dtos.WorkflowDTO, 0, len(list))
	for i := range list {
		data = append(data, *toWorkflowDTO(&list[i]))
	}
	return &dtos.WorkflowListDTO{Data: data}, nil
}

// UpdateDraft confirms ownership of a DRAFT workflow at an expected optimistic
// version, bumping the version counter (PRD §31). The draft definition body is
// kept client-side in this slice (see design D2); the operation guards against
// concurrent edits by another operator.
func (s *BuilderService) UpdateDraft(ctx context.Context, tenantID, projectID, id string, req dtos.UpdateWorkflowRequest) (*dtos.WorkflowDTO, error) {
	pid, err := s.resolveProject(ctx, tenantID, projectID)
	if err != nil {
		return nil, err
	}
	existing, err := s.workflows.FindByID(ctx, tenantID, pid, id)
	if err != nil {
		return nil, err
	}
	if existing.Status != entities.WorkflowDraft {
		return nil, domain.NewConflict("only DRAFT workflows can be edited")
	}
	if req.Version != existing.Version {
		return nil, domain.NewConflict("optimistic lock conflict: resource changed")
	}

	// The repository's optimistic update bumps version and conflicts on staleness.
	updated, err := s.workflows.UpdateStatus(ctx, tenantID, pid, id, entities.WorkflowDraft, existing.Version)
	if err != nil {
		return nil, err
	}
	return toWorkflowDTO(updated), nil
}

// Publish creates an immutable, current workflow version from the provided
// definition (PRD §3.3, §9, §55, §65, §69, §68). It uses optimistic concurrency
// on the workflow root version. On success it appends a workflow.published
// audit entry (PRD 50) attributed to the actor.
func (s *BuilderService) Publish(ctx context.Context, tenantID, projectID, id, actor string, req dtos.PublishWorkflowRequest) (*dtos.WorkflowVersionDTO, error) {
	if len(req.Definition) == 0 {
		return nil, domain.NewValidation("definition is required")
	}
	pid, err := s.resolveProject(ctx, tenantID, projectID)
	if err != nil {
		return nil, err
	}
	existing, err := s.workflows.FindByID(ctx, tenantID, pid, id)
	if err != nil {
		return nil, err
	}
	if req.Version != existing.Version {
		return nil, domain.NewConflict("optimistic lock conflict: resource changed")
	}

	versionNo := existing.CurrentVersion + 1
	version, err := s.workflows.Publish(ctx, tenantID, pid, id, versionNo, req.Definition, entities.VersionStatusPublished, existing.Version)
	if err != nil {
		return nil, err
	}
	if s.audit != nil {
		s.audit.Write(ctx, tenantID, actor, entities.AuditActionWorkflowPublished, "workflow", id,
			nil, map[string]any{"version": versionNo}, nil)
	}
	return toWorkflowVersionDTO(version), nil
}

// ListVersions returns the immutable versions of a workflow, newest first.
func (s *BuilderService) ListVersions(ctx context.Context, tenantID, projectID, id string) ([]dtos.WorkflowVersionDTO, error) {
	pid, err := s.resolveProject(ctx, tenantID, projectID)
	if err != nil {
		return nil, err
	}
	versions, err := s.workflows.ListVersions(ctx, tenantID, pid, id)
	if err != nil {
		return nil, err
	}
	out := make([]dtos.WorkflowVersionDTO, 0, len(versions))
	for i := range versions {
		out = append(out, *toWorkflowVersionDTO(&versions[i]))
	}
	return out, nil
}

// resolveProject returns the projectID to use, defaulting to the tenant's
// "default" project (creating it when missing) if none is supplied.
func (s *BuilderService) resolveProject(ctx context.Context, tenantID, projectID string) (string, error) {
	if projectID != "" {
		return projectID, nil
	}
	proj, err := s.projects.FindBySlug(ctx, tenantID, defaultProjectSlug)
	if err == nil {
		return proj.ID, nil
	}
	var de *domain.DomainError
	if !errors.As(err, &de) || de.Code != domain.ErrNotFound {
		return "", err
	}
	created, cerr := s.projects.Create(ctx, tenantID, "Default Project", defaultProjectSlug, entities.ProjectActive)
	if cerr != nil {
		return "", cerr
	}
	return created.ID, nil
}

// ---- mapping & helpers ----

func toWorkflowDTO(w *entities.Workflow) *dtos.WorkflowDTO {
	out := &dtos.WorkflowDTO{
		ID:             w.ID,
		TenantID:       w.TenantID,
		ProjectID:      w.ProjectID,
		Slug:           w.Slug,
		Name:           w.Name,
		Status:         string(w.Status),
		CurrentVersion: w.CurrentVersion,
		Version:        w.Version,
		CreatedAt:      w.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      w.UpdatedAt.Format(time.RFC3339),
	}
	if w.Description.Valid {
		out.Description = &w.Description.String
	}
	return out
}

func toWorkflowVersionDTO(v *entities.WorkflowVersion) *dtos.WorkflowVersionDTO {
	return &dtos.WorkflowVersionDTO{
		ID:         v.ID,
		WorkflowID: v.WorkflowID,
		VersionNo:  v.VersionNo,
		Status:     string(v.Status),
		IsCurrent:  v.IsCurrent,
		CreatedAt:  v.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  v.UpdatedAt.Format(time.RFC3339),
	}
}
