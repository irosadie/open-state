package entities

import (
	"time"
)

// StateInstanceStatus is the lifecycle of a runtime state occurrence (PRD §11).
type StateInstanceStatus string

const (
	StateInstanceEntering  StateInstanceStatus = "ENTERING"
	StateInstanceActive    StateInstanceStatus = "ACTIVE"
	StateInstanceWaiting   StateInstanceStatus = "WAITING"
	StateInstanceExiting   StateInstanceStatus = "EXITING"
	StateInstanceCompleted StateInstanceStatus = "COMPLETED"
	StateInstanceFailed    StateInstanceStatus = "FAILED"
	StateInstanceExpired   StateInstanceStatus = "EXPIRED"
	StateInstanceCancelled StateInstanceStatus = "CANCELLED"
)

// StateInstance is the persisted runtime occurrence of a state inside a workflow
// instance (PRD §11). It is tenant-isolated (PRD §4, §96), concurrency-safe via an
// optimistic version counter (PRD §31), and persists the retry counter (PRD §48).
// entered_at / expires_at support timeout scheduling (PRD §3.6, §25); exited_at
// records lifecycle completion.
type StateInstance struct {
	ID                 string
	TenantID           string
	WorkflowInstanceID string
	WorkflowVersionID  string
	StateKey           string // references states.key of the pinned version
	StateID            *string // optional FK to states(id)
	Status             StateInstanceStatus
	Version            int // optimistic lock (PRD §31)
	RetryCount         int // persisted retry counter (PRD §48)
	EnteredAt          time.Time
	ExpiresAt          *time.Time // state timeout (PRD §25)
	ExitedAt           *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
