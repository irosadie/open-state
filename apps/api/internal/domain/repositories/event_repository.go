package repositories

import (
	"context"
	"encoding/json"
	"time"

	"github.com/irosadie/open-state/api/internal/domain/entities"
)

// IEventRepository defines the persistence contract for the event system: append-only
// history, inbound inbox, reliable outbox, and the idempotency ledger. Every method is
// tenant-scoped: it takes an explicit tenantID (PRD §4, §96) so cross-tenant access is
// impossible at the data-access layer. It operates on domain entities (DB-agnostic,
// ADR-001) and surfaces NOT_FOUND / CONFLICT DomainErrors.
type IEventRepository interface {
	// Append persists a new immutable event to the event history (PRD §27, §51).
	Append(ctx context.Context, tenantID string, input AppendEventInput) (*entities.Event, error)
	// FindEventByID returns an event by id within a tenant.
	FindEventByID(ctx context.Context, tenantID, id string) (*entities.Event, error)
	// ListEventsByInstance returns the event history for a workflow instance in
	// deterministic sequence order (PRD §32, §52).
	ListEventsByInstance(ctx context.Context, tenantID, workflowInstanceID string) ([]entities.Event, error)
	// ListEventsByTenant returns all events for a tenant in sequence order.
	ListEventsByTenant(ctx context.Context, tenantID string) ([]entities.Event, error)

	// InsertInbox persists an inbound external event for dedup before processing (PRD §66).
	// Duplicate idempotency_key surfaces a CONFLICT.
	InsertInbox(ctx context.Context, tenantID string, input InsertInboxEventInput) (*entities.InboxEvent, error)
	// ClaimInbox atomically claims a batch of pending inbox events for processing (PRD §66).
	ClaimInbox(ctx context.Context, tenantID string, limit int) ([]entities.InboxEvent, error)
	// MarkInboxProcessed marks an inbox event PROCESSED after successful processing.
	MarkInboxProcessed(ctx context.Context, tenantID, id string) (*entities.InboxEvent, error)

	// InsertOutbox persists an event awaiting publication (PRD §65), written atomically
	// with its accompanying state change.
	InsertOutbox(ctx context.Context, tenantID string, input InsertOutboxEventInput) (*entities.OutboxEvent, error)
	// ClaimOutbox atomically claims a batch of pending outbox events for publishing (PRD §65).
	ClaimOutbox(ctx context.Context, tenantID string, limit int) ([]entities.OutboxEvent, error)
	// MarkOutboxPublished marks an outbox event PUBLISHED after successful publication.
	MarkOutboxPublished(ctx context.Context, tenantID, id string) (*entities.OutboxEvent, error)

	// UpsertIdempotency records the durable outcome of processing an idempotency key (PRD §30).
	UpsertIdempotency(ctx context.Context, tenantID string, input UpsertIdempotencyInput) (*entities.IdempotencyRecord, error)
	// FindIdempotency returns the ledger row for an idempotency key, or NOT_FOUND.
	FindIdempotency(ctx context.Context, tenantID, idempotencyKey string) (*entities.IdempotencyRecord, error)
}

// AppendEventInput carries the fields needed to append an event.
type AppendEventInput struct {
	EventID            string
	Type               string
	Source             entities.EventSource
	AggregateID        *string
	WorkflowInstanceID *string
	Timestamp          time.Time
	Payload            json.RawMessage
	CorrelationID      *string
	CausationID        *string
	IdempotencyKey     *string
}

// InsertInboxEventInput carries the fields needed to enqueue an inbound event.
type InsertInboxEventInput struct {
	IdempotencyKey string
	EventType      string
	Source         entities.EventSource
	Payload        json.RawMessage
}

// InsertOutboxEventInput carries the fields needed to enqueue an outbox event.
type InsertOutboxEventInput struct {
	EventID *string
	Payload json.RawMessage
	Topic   string
}

// UpsertIdempotencyInput carries the fields needed to record an idempotent outcome.
type UpsertIdempotencyInput struct {
	IdempotencyKey string
	Scope          string
	ResultStatus   entities.IdempotencyResultStatus
	Payload        json.RawMessage
}
