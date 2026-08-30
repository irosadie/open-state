package entities

import (
	"database/sql"
	"encoding/json"
	"time"
)

// WorkflowStatus is the lifecycle of a workflow definition (PRD §9).
type WorkflowStatus string

const (
	WorkflowDraft      WorkflowStatus = "DRAFT"
	WorkflowValidating WorkflowStatus = "VALIDATING"
	WorkflowValid      WorkflowStatus = "VALID"
	WorkflowPublished  WorkflowStatus = "PUBLISHED"
	WorkflowArchived   WorkflowStatus = "ARCHIVED"
)

// Workflow is a persisted workflow definition root, tenant+project-isolated (PRD §4, §96, §3.1.1).
type Workflow struct {
	ID              string
	TenantID        string
	ProjectID       string
	Slug            string
	Name            string
	Description     sql.NullString
	Status          WorkflowStatus
	CurrentVersion  int
	Version         int // optimistic lock (PRD §31)
	DraftDefinition json.RawMessage
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
