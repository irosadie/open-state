package entities

import "time"

// Tenant is the current tenant profile managed by the Admin Console.
type Tenant struct {
	ID          string
	Name        string
	Slug        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// TenantMembership combines a tenant role assignment with the safe user
// identity fields needed by the membership management UI.
type TenantMembership struct {
	RoleAssignmentID string
	UserID           string
	TenantID         string
	Role             UserRole
	Email            string
	Name             string
	Status           UserStatus
	Photo            *string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
