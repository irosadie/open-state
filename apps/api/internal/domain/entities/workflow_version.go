package entities

import (
	"encoding/json"
	"time"
)

// VersionStatus is the lifecycle of a workflow version snapshot (PRD §9).
type VersionStatus string

const (
	VersionStatusDraft      VersionStatus = "DRAFT"
	VersionStatusValidating VersionStatus = "VALIDATING"
	VersionStatusValid      VersionStatus = "VALID"
	VersionStatusPublished  VersionStatus = "PUBLISHED"
	VersionStatusArchived   VersionStatus = "ARCHIVED"
)

// WorkflowVersion is an immutable, versioned snapshot of a workflow definition
// (PRD §3.3, §9, §55, §58). Definition holds the full WorkflowDefinition envelope.
type WorkflowVersion struct {
	ID         string
	WorkflowID string
	TenantID   string
	ProjectID  string
	VersionNo  int
	Definition json.RawMessage // full WorkflowDefinition (PRD §68)
	Status     VersionStatus
	IsCurrent  bool // marks the active published version (PRD §58)
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
