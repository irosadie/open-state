package database

import (
	"github.com/google/uuid"
)

// Tenant scope convention (PRD 4, 96).
//
// Every repository interface method takes an explicit tenantID string — enforced
// by the compiler at the signature level — and every sqlc query filters by
// tenant_id. This file centralizes the tenant-aware helpers so the convention is
// defined once and reviewed in a single place. The helper deliberately does not
// duplicate query logic: each slice's repository already filters by tenant_id in
// its SQL. It exists to unify the naming/convention and give reviewers a single
// enforcement seam (ADR-001, portability).
//
// Future adapters (MySQL/SQLite/Mongo) MUST enforce the same tenant isolation.
type TenantScope struct {
	ID uuid.UUID
}

// NewTenantScope validates a tenant id string and returns a normalized scope. The
// caller is responsible for treating an invalid (non-UUID) tenant id as a domain
// validation error; no SQL is executed here.
func NewTenantScope(tenantID string) TenantScope {
	return TenantScope{ID: uuid.MustParse(tenantID)}
}

// TenantID returns the normalized tenant id for use as a sqlc query parameter.
func (s TenantScope) TenantID() uuid.UUID {
	return s.ID
}

// String returns the canonical tenant id as a string.
func (s TenantScope) String() string {
	return s.ID.String()
}
