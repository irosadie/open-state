package repositories

import (
	"context"
	"time"

	"github.com/irosadie/open-state/api/internal/domain/entities"
)

// IInstanceRepository defines the persistence contract for runtime workflow
// instances and their state instances. It is tenant-scoped: every method takes an
// explicit tenantID (PRD §4, §96) so cross-tenant access is impossible at the
// data-access layer. It operates on domain entities (DB-agnostic, ADR-001) and
// uses optimistic locking for concurrency safety (PRD §31).
type IInstanceRepository interface {
	// Create persists a new workflow instance against a pinned workflow version.
	Create(ctx context.Context, tenantID string, input CreateWorkflowInstanceInput) (*entities.WorkflowInstance, error)
	// FindByID returns a workflow instance by id within a tenant.
	FindByID(ctx context.Context, tenantID, id string) (*entities.WorkflowInstance, error)
	// ListByTenant returns all workflow instances for a tenant, newest first.
	ListByTenant(ctx context.Context, tenantID string) ([]entities.WorkflowInstance, error)
	// UpdateStatus updates the workflow instance status using optimistic locking (PRD §31).
	UpdateStatus(ctx context.Context, tenantID, id string, status entities.WorkflowInstanceStatus, expectedVersion int) (*entities.WorkflowInstance, error)

	// Transition atomically exits the current state instance, enters a new state
	// instance, points the workflow instance's current state at it, and increments
	// the parent workflow instance version in one transaction (PRD §69).
	Transition(ctx context.Context, tenantID string, input TransitionInput) (*entities.StateInstance, error)

	// CreateStateInstance persists a new state instance for a workflow instance.
	CreateStateInstance(ctx context.Context, tenantID string, input CreateStateInstanceInput) (*entities.StateInstance, error)
	// FindStateInstanceByID returns a state instance by id within a tenant.
	FindStateInstanceByID(ctx context.Context, tenantID, id string) (*entities.StateInstance, error)
	// UpdateStateInstanceStatus updates the state instance status using optimistic locking (PRD §31).
	UpdateStateInstanceStatus(ctx context.Context, tenantID, id string, status entities.StateInstanceStatus, expectedVersion int) (*entities.StateInstance, error)
	// IncrementRetry increments the state instance retry counter using optimistic locking (PRD §48).
	IncrementRetry(ctx context.Context, tenantID, id string, expectedVersion int) (*entities.StateInstance, error)
}

// CreateWorkflowInstanceInput carries the fields needed to create a workflow instance.
type CreateWorkflowInstanceInput struct {
	WorkflowID        string
	WorkflowVersionID string // pinned immutable version (PRD §58)
	CorrelationKey    *string
	StartedAt         *time.Time
	ExpiresAt         *time.Time
}

// CreateStateInstanceInput carries the fields needed to create a state instance.
type CreateStateInstanceInput struct {
	WorkflowInstanceID string
	WorkflowVersionID  string
	StateKey           string
	StateID            *string
}

// TransitionInput carries the data for an atomic state transition.
type TransitionInput struct {
	WorkflowInstanceID      string
	ExpectedWorkflowVersion int // optimistic lock on the parent instance (PRD §31)
	ExitStateInstanceID     string
	ExpectedExitVersion     int // optimistic lock on the exiting state instance (PRD §31)
	NewWorkflowVersionID    string
	NewStateKey             string
	NewStateID              *string
}
