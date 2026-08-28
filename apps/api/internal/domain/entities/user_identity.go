package entities

import "time"

// UserIdentity links an external OIDC provider subject to a local user (PRD §79).
// A user MAY hold one identity per provider; the (provider, subject) pair is
// globally unique.
type UserIdentity struct {
	ID              string
	UserID          string
	Provider        string
	SubjectID       string
	AutoProvisioned bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
