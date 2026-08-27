package entities

import (
	"encoding/json"
	"time"
)

// OutboxEventStatus is the delivery lifecycle of an outbound event (PRD §65).
type OutboxEventStatus string

const (
	OutboxEventPending   OutboxEventStatus = "PENDING"
	OutboxEventPublished OutboxEventStatus = "PUBLISHED"
	OutboxEventFailed    OutboxEventStatus = "FAILED"
)

// OutboxEvent is a persisted event awaiting reliable publication to the message bus
// (PRD §65). It is written atomically with the DB state change it accompanies and
// is tenant-isolated (PRD §4, §96).
type OutboxEvent struct {
	ID           string
	TenantID     string
	EventID      *string // reference to the originating events row (PRD §27)
	Payload      json.RawMessage
	Topic        string // destination topic on the bus
	Status       OutboxEventStatus
	AttemptCount int // publish retry counter (PRD §65)
	PublishedAt  *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
