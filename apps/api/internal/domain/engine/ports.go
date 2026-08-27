package engine

import "context"

// WorkflowRepository provides access to workflow definitions.
// Implemented by a concrete adapter (PostgreSQL) in Epic #3.
type WorkflowRepository interface {
	// GetBySlug returns a published workflow definition by slug.
	GetBySlug(ctx context.Context, tenantID, slug string) (*WorkflowDefinition, error)
	// Save persists a workflow definition.
	Save(ctx context.Context, def *WorkflowDefinition) error
}

// InstanceRepository provides access to workflow instances.
// Implemented by a concrete adapter (PostgreSQL) in Epic #3.
type InstanceRepository interface {
	// Create persists a new workflow instance.
	Create(ctx context.Context, instance *WorkflowInstance) error
	// Get returns a workflow instance by id.
	Get(ctx context.Context, tenantID, instanceID string) (*WorkflowInstance, error)
	// UpdateWithVersion updates an instance, enforcing optimistic concurrency.
	// Returns ErrConflict if the stored version != expectedVersion.
	UpdateWithVersion(ctx context.Context, instance *WorkflowInstance, expectedVersion int) error
}

// EventRepository provides access to runtime events and idempotency tracking.
// Implemented by a concrete adapter (PostgreSQL) in Epic #3.
type EventRepository interface {
	// Append persists a new event (append-only).
	Append(ctx context.Context, event *Event) error
	// IsProcessed reports whether an idempotency key was already processed.
	IsProcessed(ctx context.Context, tenantID, idempotencyKey string) (bool, error)
	// MarkProcessed records an idempotency key as processed.
	MarkProcessed(ctx context.Context, tenantID, idempotencyKey, eventID string) error
}

// EngineRepositories bundles the ports the engine depends on.
type EngineRepositories struct {
	Workflows WorkflowRepository
	Instances InstanceRepository
	Events    EventRepository
}
