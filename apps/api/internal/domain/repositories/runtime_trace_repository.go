package repositories

import (
	"context"
	"time"

	"github.com/irosadie/open-state/api/internal/domain/entities"
)

// IRuntimeTraceRepository is the tenant-scoped append-only trace port. There
// are intentionally no update or delete methods.
type IRuntimeTraceRepository interface {
	Append(ctx context.Context, tenantID string, input AppendRuntimeTraceInput) (*entities.RuntimeTraceEntry, error)
	ListByInstance(ctx context.Context, tenantID, workflowInstanceID string) ([]entities.RuntimeTraceEntry, error)
	ListByTurn(ctx context.Context, tenantID, workflowInstanceID, turnID string) ([]entities.RuntimeTraceEntry, error)
}

// AppendRuntimeTraceInput contains only application-observed trace fields.
type AppendRuntimeTraceInput struct {
	WorkflowInstanceID string
	TurnID             *string
	Stage              entities.RuntimeTraceStage
	Source             entities.RuntimeTraceSource
	Status             entities.RuntimeTraceStatus
	OccurredAt         time.Time
	CorrelationID      *string
	DurationMS         *int64
	ReasonCode         *string
	ErrorCode          *string
	ProviderAlias      *string
	ProviderReference  *string
	Summary            *string
	Attributes         entities.SanitizedAttributes
}

// IRuntimeReadRepository provides the definition and state projections needed
// by the inspector without expanding the command-oriented repository ports.
type IRuntimeReadRepository interface {
	FindWorkflow(ctx context.Context, tenantID, workflowID string) (*entities.Workflow, error)
	FindWorkflowVersion(ctx context.Context, tenantID, workflowVersionID string) (*entities.WorkflowVersion, error)
	ListStatesByVersion(ctx context.Context, tenantID, workflowVersionID string) ([]entities.State, error)
	ListStateInstancesByWorkflowInstance(ctx context.Context, tenantID, workflowInstanceID string) ([]entities.StateInstance, error)
}
