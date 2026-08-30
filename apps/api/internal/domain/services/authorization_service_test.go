package services

import (
	"testing"

	"github.com/irosadie/open-state/api/internal/domain/entities"
)

func TestPermissionsForRole(t *testing.T) {
	owner := PermissionsForRole(entities.UserRoleOwner)
	viewer := PermissionsForRole(entities.UserRoleViewer)

	// OWNER holds user:* and tenant:* (sole holder).
	if !contains(owner, "user:*") {
		t.Error("expected OWNER to hold user:*")
	}
	if !contains(owner, "tenant:*") {
		t.Error("expected OWNER to hold tenant:*")
	}
	if contains(viewer, "user:*") {
		t.Error("expected VIEWER to NOT hold user:*")
	}
}

func TestHasPermission(t *testing.T) {
	cases := []struct {
		name     string
		role     entities.UserRole
		required Permission
		want     bool
	}{
		{"owner workflow publish", entities.UserRoleOwner, "workflow:publish", true},
		{"owner user manage", entities.UserRoleOwner, "user:manage", true},
		{"admin workflow publish", entities.UserRoleAdmin, "workflow:publish", true},
		{"admin user manage denied", entities.UserRoleAdmin, "user:manage", false},
		{"editor workflow create", entities.UserRoleEditor, "workflow:create", true},
		{"editor capability delete denied", entities.UserRoleEditor, "capability:delete", false},
		{"viewer workflow read", entities.UserRoleViewer, "workflow:read", true},
		{"viewer workflow create denied", entities.UserRoleViewer, "workflow:create", false},
		{"operator instance retry", entities.UserRoleOperator, "instance:retry", true},
		{"operator debug read", entities.UserRoleOperator, "debug:read", true},
		{"operator instance read", entities.UserRoleOperator, "instance:read", true},
		{"viewer debug read denied", entities.UserRoleViewer, "debug:read", false},
		{"admin instance read", entities.UserRoleAdmin, "instance:read", true},
		{"unknown role denied", entities.UserRole("NOPE"), "workflow:read", false},
		{"audit read viewer", entities.UserRoleViewer, "audit:read", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := HasPermission(tc.role, tc.required); got != tc.want {
				t.Errorf("HasPermission(%s, %s) = %v, want %v", tc.role, tc.required, got, tc.want)
			}
		})
	}
}

func TestWildcardMatches(t *testing.T) {
	if !wildcardMatches("workflow:*", "workflow:publish") {
		t.Error("expected wildcard to match concrete permission")
	}
	if wildcardMatches("workflow:*", "instance:read") {
		t.Error("expected wildcard NOT to match unrelated resource")
	}
	if wildcardMatches("workflow:*", "workflow") {
		t.Error("expected wildcard NOT to match a prefix without colon-suffix")
	}
}

func contains(perms []Permission, want Permission) bool {
	for _, p := range perms {
		if p == want {
			return true
		}
	}
	return false
}
