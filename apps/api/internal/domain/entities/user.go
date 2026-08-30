package entities

import "time"

type UserRole string
type UserStatus string

const (
	// UserRoleOwner, UserRoleAdmin, UserRoleEditor, UserRoleOperator, and
	// UserRoleViewer are the tenant-scoped roles defined in PRD 80.
	// The effective role for authorization is read from role_assignments
	// (PRD 81), not from the deprecated users.role column.
	UserRoleOwner    UserRole = "OWNER"
	UserRoleAdmin    UserRole = "ADMIN"
	UserRoleEditor   UserRole = "EDITOR"
	UserRoleOperator UserRole = "OPERATOR"
	UserRoleViewer   UserRole = "VIEWER"

	// UserRoleLegacy is the deprecated value written to the legacy users.role
	// column (a PostgreSQL ENUM accepting only 'USER'/'ADMIN'). It is NOT an
	// effective role: authorization reads from role_assignments (PRD 81), and
	// an absent role assignment defaults to the least-privilege UserRoleViewer.
	UserRoleLegacy UserRole = "USER"

	UserStatusActive    UserStatus = "ACTIVE"
	UserStatusSuspended UserStatus = "SUSPENDED"
)

// TenantRoles is the complete role set accepted by tenant membership
// administration. Keeping this list in the domain prevents arbitrary strings
// from reaching role_assignments.
var TenantRoles = []UserRole{
	UserRoleOwner,
	UserRoleAdmin,
	UserRoleEditor,
	UserRoleOperator,
	UserRoleViewer,
}

func IsTenantRole(role UserRole) bool {
	for _, allowed := range TenantRoles {
		if role == allowed {
			return true
		}
	}
	return false
}

type User struct {
	ID           string
	Email        string
	PasswordHash string
	Name         string
	Role         UserRole
	Status       UserStatus
	Photo        *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// RoleAssignment is a user's tenant-scoped role (PRD 81). A user may hold
// different roles in different tenants; the (user_id, tenant_id) pair is unique.
type RoleAssignment struct {
	ID        string
	UserID    string
	TenantID  string
	Role      UserRole
	CreatedAt time.Time
	UpdatedAt time.Time
}
