package entities

import (
	"encoding/json"
	"time"
)

// IdempotencyResultStatus is the durable outcome of processing an idempotency key
// (PRD §30).
type IdempotencyResultStatus string

const (
	IdempotencyResultProcessed IdempotencyResultStatus = "PROCESSED"
	IdempotencyResultIgnored   IdempotencyResultStatus = "IGNORED"
	IdempotencyResultFailed    IdempotencyResultStatus = "FAILED"
)

// IdempotencyRecord is a tenant-scoped dedup ledger row keyed by idempotency_key,
// storing the result so repeated external deliveries are skipped (PRD §30).
type IdempotencyRecord struct {
	ID             string
	TenantID       string
	IdempotencyKey string // dedup key, unique per tenant (PRD §30)
	Scope          string // record scope, defaults to "event"
	ResultStatus   IdempotencyResultStatus
	Payload        json.RawMessage // optional outcome payload
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
