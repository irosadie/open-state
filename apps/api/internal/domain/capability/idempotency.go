package capability

import (
	"context"
	"sync"
)

// IdempotencyStore is a port for deduplicating capability invocations with
// side effects (PRD §64). The concrete implementation may be in-memory (tests)
// or backed by PostgreSQL later.
type IdempotencyStore interface {
	// Get returns a previously stored result for an idempotency key.
	Get(ctx context.Context, key string) (*InvocationResult, bool, error)
	// Put stores the result for an idempotency key.
	Put(ctx context.Context, key string, result InvocationResult) error
}

// InMemoryIdempotencyStore is a thread-safe in-memory implementation used for
// tests and default (non-persistent) operation.
type InMemoryIdempotencyStore struct {
	mu      sync.RWMutex
	results map[string]InvocationResult
}

// NewInMemoryIdempotencyStore constructs an empty store.
func NewInMemoryIdempotencyStore() *InMemoryIdempotencyStore {
	return &InMemoryIdempotencyStore{results: map[string]InvocationResult{}}
}

// Get implements IdempotencyStore.
func (s *InMemoryIdempotencyStore) Get(_ context.Context, key string) (*InvocationResult, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res, ok := s.results[key]
	if !ok {
		return nil, false, nil
	}
	cp := res
	return &cp, true, nil
}

// Put implements IdempotencyStore.
func (s *InMemoryIdempotencyStore) Put(_ context.Context, key string, result InvocationResult) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.results[key] = result
	return nil
}

// buildIdempotencyKey derives the idempotency key from workflow instance +
// action id (PRD §64). Returns empty if either is missing.
func buildIdempotencyKey(workflowInstanceID, actionID string) string {
	if workflowInstanceID == "" || actionID == "" {
		return ""
	}
	return workflowInstanceID + ":" + actionID
}
