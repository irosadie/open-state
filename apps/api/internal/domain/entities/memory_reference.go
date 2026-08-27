package entities

import (
	"encoding/json"
	"time"
)

// MemoryReference is persistent user/customer memory (PRD §24, §43.2). It is
// distinct from workflow data: deleting a workflow instance never deletes memory.
// SourceWorkflowInstanceID is optional provenance and is a soft reference, not a
// hard FK, so memory survives instance expiry/deletion (PRD §24).
type MemoryReference struct {
	ID                       string
	TenantID                 string
	OwnerType                string          // e.g. CUSTOMER/USER
	OwnerID                  string
	Name                     string          // e.g. address/preferences
	Value                    json.RawMessage // stored value / snapshot
	SourceWorkflowInstanceID *string         // optional provenance, soft reference
	CreatedAt                time.Time
	UpdatedAt                time.Time
}
