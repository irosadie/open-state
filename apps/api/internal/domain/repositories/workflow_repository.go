package repositories

import (
	"context"

	"github.com/irosadie/open-state/api/internal/domain/entities"
)

// IWorkflowRepository defines the persistence contract for workflow definitions,
// versions, states, transitions, and guards. It is tenant+project-scoped: every
// method takes explicit tenantID and projectID (PRD §4, §96, §3.1.1) so
// cross-tenant/project access is impossible at the data-access layer. It operates
// on domain entities (DB-agnostic, ADR-001).
type IWorkflowRepository interface {
	// Create persists a new workflow definition root within a project.
	Create(ctx context.Context, tenantID, projectID, slug, name string, description *string, draftDefinition []byte) (*entities.Workflow, error)
	// FindByID returns a workflow by id within a tenant and project.
	FindByID(ctx context.Context, tenantID, projectID, id string) (*entities.Workflow, error)
	// FindBySlug returns a workflow by slug within a tenant and project (PRD §5).
	FindBySlug(ctx context.Context, tenantID, projectID, slug string) (*entities.Workflow, error)
	// ListByTenant returns all workflows for a tenant within a project.
	ListByTenant(ctx context.Context, tenantID, projectID string) ([]entities.Workflow, error)
	// UpdateStatus updates workflow status using optimistic concurrency (PRD §31).
	UpdateStatus(ctx context.Context, tenantID, projectID, id string, status entities.WorkflowStatus, expectedVersion int) (*entities.Workflow, error)
	// UpdateDraft atomically replaces the editable graph and metadata under an
	// optimistic lock, incrementing the workflow version on success.
	UpdateDraft(ctx context.Context, tenantID, projectID, id, name string, description *string, draftDefinition []byte, expectedVersion int) (*entities.Workflow, error)

	// CreateVersion persists a new immutable workflow version snapshot.
	CreateVersion(ctx context.Context, tenantID, projectID, workflowID string, versionNo int, definition []byte, status entities.VersionStatus, isCurrent bool) (*entities.WorkflowVersion, error)
	// Publish atomically inserts a new version, marks it current, and bumps the
	// workflow current_version + optimistic version counter in one transaction (PRD §9, §65, §69).
	Publish(ctx context.Context, tenantID, projectID, workflowID string, versionNo int, definition []byte, status entities.VersionStatus, expectedVersion int) (*entities.WorkflowVersion, error)
	// FindCurrentVersion returns the active (is_current) version of a workflow (PRD §58).
	FindCurrentVersion(ctx context.Context, tenantID, projectID, workflowID string) (*entities.WorkflowVersion, error)
	// FindCurrentVersionByWorkflow returns the active version of a workflow by workflowID+tenantID only.
	FindCurrentVersionByWorkflow(ctx context.Context, tenantID, workflowID string) (*entities.WorkflowVersion, error)
	// ListVersions returns all versions of a workflow, newest first.
	ListVersions(ctx context.Context, tenantID, projectID, workflowID string) ([]entities.WorkflowVersion, error)
	// FindVersionByNumber returns a specific version of a workflow.
	FindVersionByNumber(ctx context.Context, tenantID, projectID, workflowID string, versionNo int) (*entities.WorkflowVersion, error)

	// CreateState persists a relational state projection for a workflow version.
	CreateState(ctx context.Context, tenantID, projectID, workflowVersionID, key string, kind entities.StateKind, name string, description, instructions *string, requiredContext, capabilities, policy, position []byte, isTerminal bool) (*entities.State, error)
	// ListStatesByVersion returns all states for a workflow version.
	ListStatesByVersion(ctx context.Context, tenantID, projectID, workflowVersionID string) ([]entities.State, error)

	// CreateTransition persists a relational transition for a workflow version.
	CreateTransition(ctx context.Context, tenantID, projectID, workflowVersionID, key, sourceStateID, targetStateID, event string, priority int, isActive bool) (*entities.Transition, error)
	// ListTransitionsByVersion returns all transitions for a workflow version, ordered by priority (PRD §34).
	ListTransitionsByVersion(ctx context.Context, tenantID, projectID, workflowVersionID string) ([]entities.Transition, error)

	// CreateTransitionGuard persists a guard group for a transition.
	CreateTransitionGuard(ctx context.Context, tenantID, projectID, transitionID, workflowVersionID, logic string, conditions []byte) (*entities.TransitionGuard, error)
	// ListGuardsByTransition returns all guards for a transition.
	ListGuardsByTransition(ctx context.Context, tenantID, projectID, transitionID string) ([]entities.TransitionGuard, error)
}
