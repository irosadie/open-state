package entities

import (
	"database/sql"
	"encoding/json"
	"time"
)

// StateKind identifies the logical role of a state node in a workflow graph (PRD §14).
type StateKind string

const (
	StateKindStart    StateKind = "START"
	StateKindState    StateKind = "STATE"
	StateKindDecision StateKind = "DECISION"
	StateKindWait     StateKind = "WAIT"
	StateKindEnd      StateKind = "END"
	StateKindEvent    StateKind = "EVENT"
)

// State is a persisted, relational projection of a workflow definition node
// scoped to an immutable workflow version (PRD §12, §14).
type State struct {
	ID                string
	WorkflowVersionID string
	Key               string // stable node key, e.g. PAYMENT
	Kind              StateKind
	Name              string
	Description       sql.NullString
	Instructions      sql.NullString
	RequiredContext   json.RawMessage // JSON array of context keys
	Capabilities      json.RawMessage // JSON array of capability refs
	Policy            json.RawMessage // JSON StatePolicy
	IsTerminal        bool
	Position          json.RawMessage // JSON x/y for builder
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
