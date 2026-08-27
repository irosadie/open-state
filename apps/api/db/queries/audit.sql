-- Append-only audit trail (PRD 50). All queries are tenant-scoped (PRD 4, 96).
-- No UPDATE/DELETE statements exist for audit_logs: entries are append-only.

-- name: AppendAuditLog :one
INSERT INTO audit_logs (tenant_id, actor, action, resource_type, resource_id, before, after, correlation_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, tenant_id, actor, action, resource_type, resource_id, before, after, correlation_id, occurred_at, created_at;

-- name: ListAuditByTenant :many
SELECT id, tenant_id, actor, action, resource_type, resource_id, before, after, correlation_id, occurred_at, created_at
FROM audit_logs
WHERE tenant_id = $1
ORDER BY occurred_at DESC;

-- name: ListAuditByAction :many
SELECT id, tenant_id, actor, action, resource_type, resource_id, before, after, correlation_id, occurred_at, created_at
FROM audit_logs
WHERE tenant_id = $1 AND action = $2
ORDER BY occurred_at DESC;

-- name: ListAuditByResource :many
SELECT id, tenant_id, actor, action, resource_type, resource_id, before, after, correlation_id, occurred_at, created_at
FROM audit_logs
WHERE tenant_id = $1 AND resource_type = $2 AND resource_id = $3
ORDER BY occurred_at DESC;
