package repositories

import (
	"context"

	"github.com/irosadie/open-state/api/internal/domain/entities"
)

// IContextRepository defines the persistence contract for scoped runtime context
// (context_records) and persistent memory references (memory_references).
// Tenant-scoped: every method takes an explicit tenantID (PRD §4, §96). Operates
// on domain entities (ADR-001) and returns DomainError NOT_FOUND / CONFLICT
// (PRD §24, §31). Distinct concerns: context is scoped/versioned runtime data;
// memory is long-lived user/customer data that survives workflow expiry.
type IContextRepository interface {
	// UpsertContext inserts a new context value for a scope, or updates an existing
	// one (version bumped) when its current version matches expectedVersion (optimistic
	// concurrency, PRD §31). A stale version returns CONFLICT.
	UpsertContext(ctx context.Context, tenantID string, scopeType entities.ContextScopeType, scopeID, key string, value []byte, expectedVersion int) (*entities.ContextRecord, error)
	// FindContextByScope returns a single context value for a scope and key.
	FindContextByScope(ctx context.Context, tenantID string, scopeType entities.ContextScopeType, scopeID, key string) (*entities.ContextRecord, error)
	// ListContextByScope returns the full context snapshot for a scope.
	ListContextByScope(ctx context.Context, tenantID string, scopeType entities.ContextScopeType, scopeID string) ([]entities.ContextRecord, error)
	// DeleteContext removes a single context value for a scope and key.
	DeleteContext(ctx context.Context, tenantID string, scopeType entities.ContextScopeType, scopeID, key string) error

	// UpsertMemoryReference inserts or updates a persistent memory reference for an
	// owner. sourceWorkflowInstanceID is optional provenance (PRD §24).
	UpsertMemoryReference(ctx context.Context, tenantID, ownerType, ownerID, name string, value []byte, sourceWorkflowInstanceID *string) (*entities.MemoryReference, error)
	// FindMemoryReference returns a single persistent memory reference for an owner.
	FindMemoryReference(ctx context.Context, tenantID, ownerType, ownerID, name string) (*entities.MemoryReference, error)
	// ListMemoryByOwner returns all persistent memory for an owner.
	ListMemoryByOwner(ctx context.Context, tenantID, ownerType, ownerID string) ([]entities.MemoryReference, error)
	// DeleteMemoryReference removes a single persistent memory reference for an owner.
	DeleteMemoryReference(ctx context.Context, tenantID, ownerType, ownerID, name string) error
}
