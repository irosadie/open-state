package entities

import (
	"database/sql"
	"time"
)

// WorkflowInstanceStatus is the lifecycle of a running workflow instance (PRD §10, §42-43).
type WorkflowInstanceStatus string

const (
	WorkflowInstanceCreated    WorkflowInstanceStatus = "CREATED"
	WorkflowInstanceRunning    WorkflowInstanceStatus = "RUNNING"
	WorkflowInstanceWaiting    WorkflowInstanceStatus = "WAITING"
	WorkflowInstanceCompleted  WorkflowInstanceStatus = "COMPLETED"
	WorkflowInstanceCancelled  WorkflowInstanceStatus = "CANCELLED"
	WorkflowInstanceFailed     WorkflowInstanceStatus = "FAILED"
	WorkflowInstanceExpired    WorkflowInstanceStatus = "EXPIRED"
	WorkflowInstanceAborted    WorkflowInstanceStatus = "ABORTED"
	WorkflowInstanceSuspended  WorkflowInstanceStatus = "SUSPENDED"
)

// WorkflowInstance is the persisted runtime execution root: a running copy of a
// published workflow version, pinned to an immutable workflow_version_id for
// reproducibility (PRD §10, §58). It is tenant-isolated (PRD §4, §96) and
// concurrency-safe via an optimistic version counter (PRD §31).
type WorkflowInstance struct {
	ID                    string
	TenantID              string
	WorkflowID            string
	WorkflowVersionID     string
	CorrelationKey        sql.NullString // business/conversation correlation (PRD §6)
	Status                WorkflowInstanceStatus
	Version               int // optimistic lock (PRD §31)
	CurrentStateInstanceID *string // active state instance for fast "current state" resolution (PRD §7)
	StartedAt             *time.Time
	CompletedAt           *time.Time
	ExpiresAt             *time.Time // workflow timeout (PRD §26)
	CreatedAt             time.Time
	UpdatedAt             time.Time
}
