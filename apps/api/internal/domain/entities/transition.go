package entities

import (
	"encoding/json"
	"time"
)

// Transition is a valid movement from one state to another based on an event
// and optional guards (PRD §33). Immutable per workflow version.
type Transition struct {
	ID                string
	WorkflowVersionID string
	Key               string // stable transition id
	SourceStateID     string
	TargetStateID     string
	Event             string
	Priority          int // lower value = evaluated first (PRD §34)
	IsActive          bool
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// TransitionGuard is a deterministic guard condition group attached to a
// transition (PRD §35).
type TransitionGuard struct {
	ID                string
	TransitionID      string
	WorkflowVersionID string
	Logic             string          // AND/OR
	Conditions        json.RawMessage // JSON array of guard conditions
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
