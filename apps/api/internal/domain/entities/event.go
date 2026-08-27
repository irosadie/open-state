package entities

import (
	"database/sql"
	"encoding/json"
	"time"
)

// EventSource is the origin of an event (PRD §28).
type EventSource string

const (
	EventSourceUser      EventSource = "USER"
	EventSourceLLM       EventSource = "LLM"
	EventSourceMCP       EventSource = "MCP"
	EventSourceWebhook   EventSource = "WEBHOOK"
	EventSourceSystem    EventSource = "SYSTEM"
	EventSourceScheduler EventSource = "SCHEDULER"
	EventSourceAdmin     EventSource = "ADMIN"
	EventSourceAPI       EventSource = "API"
)

// Event is an immutable, append-only event-history record (PRD §27, §51). It carries
// the full event model and is tenant-isolated (PRD §4, §96). Events are ordered per
// workflow instance by a monotonic sequence (PRD §32).
type Event struct {
	ID                 string
	TenantID           string
	EventID            string          // logical id, unique per tenant (PRD §27)
	Type               string          // e.g. payment.success
	Source             EventSource     // origin (PRD §28)
	AggregateID        sql.NullString  // aggregate this event refers to
	WorkflowInstanceID *string         // workflow instance this event belongs to (PRD §32)
	Sequence           int64           // monotonic per-tenant ordering (PRD §32)
	Timestamp          time.Time       // event time (PRD §27)
	Payload            json.RawMessage // typed event payload (PRD §27)
	CorrelationID      sql.NullString  // conversation/business correlation (PRD §27)
	CausationID        sql.NullString  // causally prior event (PRD §27)
	IdempotencyKey     sql.NullString  // dedup key (PRD §30)
	CreatedAt          time.Time
}
