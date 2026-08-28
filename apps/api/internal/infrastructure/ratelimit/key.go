package ratelimit

import "fmt"

// Scope key helpers build consistent, documented keys for rate limiting
// (PRD §83). Keys use a `dimension:value` format so limits can be scoped by
// tenant, user, API key, or operation without collisions.

// TenantKey scopes a limit to a tenant.
func TenantKey(tenantID string) string { return "tenant:" + tenantID }

// UserKey scopes a limit to a user.
func UserKey(userID string) string { return "user:" + userID }

// APIKey scopes a limit to an API key.
func APIKey(key string) string { return "apikey:" + key }

// RouteUserKey scopes a limit to a specific operation for a user (e.g. login).
func RouteUserKey(route, userID string) string { return fmt.Sprintf("route:%s:user:%s", route, userID) }

// RouteIPKey scopes a limit to a specific operation for a client IP
// (e.g. register, where no account identity exists yet).
func RouteIPKey(route, ip string) string { return fmt.Sprintf("route:%s:ip:%s", route, ip) }

// TenantCapabilityKey scopes a limit to a capability within a tenant.
func TenantCapabilityKey(tenantID, capabilityID string) string {
	return fmt.Sprintf("tenant:%s:capability:%s", tenantID, capabilityID)
}
