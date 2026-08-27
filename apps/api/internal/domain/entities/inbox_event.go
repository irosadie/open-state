package entities

import (
	"encoding/json"
	"time"
)

// InboxEventStatus is the processing lifecycle of an inbound external event (PRD §66).
type InboxEventStatus string

const (
	InboxEventReceived   InboxEventStatus = "RECEIVED"
	InboxEventProcessing InboxEventStatus = "PROCESSING"
	InboxEventProcessed  InboxEventStatus = "PROCESSED"
	InboxEventFailed     InboxEventStatus = "FAILED"
)

// InboxEvent is a persisted inbound external event queued for processing and
// deduplicated by idempotency_key (PRD §66, §30). It is tenant-isolated (PRD §4, §96).
type InboxEvent struct {
	ID             string
	TenantID       string
	IdempotencyKey string          // dedup key, unique per tenant (PRD §30, §66)
	EventType      string          // e.g. payment.success
	Source         EventSource     // origin (PRD §28)
	Payload        json.RawMessage // inbound payload (PRD §27)
	Status         InboxEventStatus
	AttemptCount   int // retry counter (PRD §66)
	ReceivedAt     time.Time
	ProcessedAt    *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
