package services

import (
	"github.com/irosadie/open-state/api/internal/domain/entities"
)

// Permission is a fine-grained action a user may perform on a resource
// (e.g. "workflow:create", "capability:delete"). The role-permission matrix
// (PRD 80) maps each role to its permitted set of permissions.
type Permission string

// RolePermissionMatrix maps each tenant role (PRD 80) to the set of permissions
// it grants. OWNER is the only role granted user:* and tenant:* permissions.
// Wildcards (e.g. "workflow:*") are matched wildcard-first: they cover every
// concrete "workflow:<verb>" permission.
var RolePermissionMatrix = map[entities.UserRole][]Permission{
	entities.UserRoleOwner: {
		"workflow:*", "capability:*", "binding:*", "user:*", "audit:*", "tenant:*", "instance:*", "debug:*", "api_key:*",
	},
	entities.UserRoleAdmin: {
		"workflow:*", "capability:*", "binding:*", "audit:*", "instance:*", "debug:*", "api_key:*",
	},
	entities.UserRoleEditor: {
		"workflow:read", "workflow:create", "workflow:update", "workflow:publish", "workflow:simulate",
		"capability:read", "binding:read",
	},
	entities.UserRoleOperator: {
		"instance:read", "instance:retry", "instance:suspend", "instance:resume", "debug:read",
		"workflow:read", "capability:read", "binding:read",
	},
	entities.UserRoleViewer: {
		"workflow:read", "capability:read", "binding:read", "instance:read", "audit:read",
	},
}

// PermissionsForRole returns the granted permissions for a role. An unknown role
// yields an empty set (default deny).
func PermissionsForRole(role entities.UserRole) []Permission {
	return RolePermissionMatrix[role]
}

// HasPermission reports whether the role grants the required permission,
// wildcard-first (e.g. "workflow:*" matches "workflow:publish").
func HasPermission(role entities.UserRole, required Permission) bool {
	for _, p := range RolePermissionMatrix[role] {
		if p == required || (len(p) > 2 && p[len(p)-1] == '*' && wildcardMatches(string(p), string(required))) {
			return true
		}
	}
	return false
}

// wildcardMatches reports whether a wildcard permission (e.g. "workflow:*")
// covers a concrete permission (e.g. "workflow:create").
func wildcardMatches(wildcard, concrete string) bool {
	prefix := wildcard[:len(wildcard)-1] // strip trailing '*'
	return len(concrete) > len(prefix) && concrete[:len(prefix)] == prefix
}
