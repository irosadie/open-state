package entities

import "time"

// ProjectStatus is the lifecycle of a project (PRD §3.1.1).
type ProjectStatus string

const (
	ProjectActive   ProjectStatus = "ACTIVE"
	ProjectArchived ProjectStatus = "ARCHIVED"
)

// Project is a business area owned by a tenant (PRD §3.1.1).
// A tenant owns many projects; each project owns many workflows.
type Project struct {
	ID        string
	TenantID  string
	Name      string
	Slug      string
	Status    ProjectStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}
