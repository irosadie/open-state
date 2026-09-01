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
	workflows           repositories.IWorkflowRepository
	projects            repositories.IProjectRepository
	audit               *AuditWriter
	mcpBindingValidator MCPBindingWorkflowValidator
	now                 func() time.Time
}

// MCPBindingWorkflowValidator is the narrow authoring contract used by the
// builder to reject workflows whose required MCP bindings are unavailable.
type MCPBindingWorkflowValidator interface {
	ValidateWorkflow(ctx context.Context, tenantID, projectID string, definition []byte) (*domain.DomainError, error)
}

// NewBuilderService builds a BuilderService.
func NewBuilderService(workflows repositories.IWorkflowRepository, projects repositories.IProjectRepository, audit *AuditWriter, validators ...MCPBindingWorkflowValidator) *BuilderService {
	var validator MCPBindingWorkflowValidator
	if len(validators) > 0 {
		validator = validators[0]
	}
	return &BuilderService{workflows: workflows, projects: projects, audit: audit, mcpBindingValidator: validator, now: time.Now}
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
	if validationErr := ensureWorkflowDefinitionJSON(req.Definition); validationErr != nil {
		return nil, validationErr
	}

	projectID, err := s.resolveProject(ctx, tenantID, req.ProjectID)
	if err != nil {
		return nil, err
	}
	if validationErr, validationErrCause := s.validateMCPBindings(ctx, tenantID, projectID, req.Definition); validationErrCause != nil {
		return nil, validationErrCause
	} else if validationErr != nil {
		return nil, validationErr
	}

	wf, err := s.workflows.Create(ctx, tenantID, projectID, req.Slug, req.Name, optional(req.Description), req.Definition)
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

// UpdateDraft confirms ownership of an editable workflow at an expected
// optimistic version and atomically persists its metadata and graph body.
func (s *BuilderService) UpdateDraft(ctx context.Context, tenantID, projectID, id string, req dtos.UpdateWorkflowRequest) (*dtos.WorkflowDTO, error) {
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
	if validationErr := ensureWorkflowDefinitionJSON(req.Definition); validationErr != nil {
		return nil, validationErr
	}
	if validationErr, validationErrCause := s.validateMCPBindings(ctx, tenantID, pid, req.Definition); validationErrCause != nil {
		return nil, validationErrCause
	} else if validationErr != nil {
		return nil, validationErr
	}

	description := req.Description
	if description == nil && existing.Description.Valid {
		description = &existing.Description.String
	}
	updated, err := s.workflows.UpdateDraft(ctx, tenantID, pid, id, req.Name, description, req.Definition, existing.Version)
	if err != nil {
		return nil, err
	}
	return toWorkflowDTO(updated), nil
}

// Publish validates and creates an immutable, current workflow version from the
// persisted draft definition (PRD §3.3, §9, §55, §65, §69, §68).
func (s *BuilderService) Publish(ctx context.Context, tenantID, projectID, id, actor string, req dtos.PublishWorkflowRequest) (*dtos.WorkflowVersionDTO, error) {
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
	if validationErr := validateWorkflowDefinition(existing.DraftDefinition); validationErr != nil {
		return nil, validationErr
	}
	if validationErr, validationErrCause := s.validateMCPBindings(ctx, tenantID, pid, existing.DraftDefinition); validationErrCause != nil {
		return nil, validationErrCause
	} else if validationErr != nil {
		return nil, validationErr
	}

	versionNo := existing.CurrentVersion + 1
	version, err := s.workflows.Publish(ctx, tenantID, pid, id, versionNo, existing.DraftDefinition, entities.VersionStatusPublished, existing.Version)
	if err != nil {
		return nil, err
	}
	if s.audit != nil {
		s.audit.Write(ctx, tenantID, actor, entities.AuditActionWorkflowPublished, "workflow", id,
			nil, map[string]any{"version": versionNo}, nil)
	}
	return toWorkflowVersionDTO(version), nil
}

// GetVersion returns one immutable published snapshot within the workflow scope.
func (s *BuilderService) GetVersion(ctx context.Context, tenantID, projectID, id string, versionNo int) (*dtos.WorkflowVersionDTO, error) {
	pid, err := s.resolveProject(ctx, tenantID, projectID)
	if err != nil {
		return nil, err
	}
	version, err := s.workflows.FindVersionByNumber(ctx, tenantID, pid, id, versionNo)
	if err != nil {
		return nil, err
	}
	return toWorkflowVersionDTO(version), nil
}

// CompareVersions compares two immutable snapshots of one workflow.
func (s *BuilderService) CompareVersions(ctx context.Context, tenantID, projectID, id string, baseVersion, targetVersion int) (*dtos.WorkflowDiffDTO, error) {
	if baseVersion < 1 || targetVersion < 1 || baseVersion == targetVersion {
		return nil, domain.NewValidation("baseVersion and targetVersion must be distinct positive versions")
	}
	pid, err := s.resolveProject(ctx, tenantID, projectID)
	if err != nil {
		return nil, err
	}
	base, err := s.workflows.FindVersionByNumber(ctx, tenantID, pid, id, baseVersion)
	if err != nil {
		return nil, err
	}
	target, err := s.workflows.FindVersionByNumber(ctx, tenantID, pid, id, targetVersion)
	if err != nil {
		return nil, err
	}
	return compareWorkflowDefinitions(id, baseVersion, targetVersion, base.Definition, target.Definition)
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
		project, err := s.projects.FindByID(ctx, tenantID, projectID)
		if err != nil {
			return "", err
		}
		return project.ID, nil
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

func (s *BuilderService) validateMCPBindings(ctx context.Context, tenantID, projectID string, definition []byte) (*domain.DomainError, error) {
	if s.mcpBindingValidator == nil {
		return nil, nil
	}
	return s.mcpBindingValidator.ValidateWorkflow(ctx, tenantID, projectID, definition)
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
		Definition:     w.DraftDefinition,
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
		Definition: v.Definition,
	}
}
