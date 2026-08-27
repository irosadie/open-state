-- name: CreateWorkflowInstance :one
INSERT INTO workflow_instances (tenant_id, workflow_id, workflow_version_id, correlation_key, status, started_at, expires_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, tenant_id, workflow_id, workflow_version_id, correlation_key, status, version, current_state_instance_id, started_at, completed_at, expires_at, created_at, updated_at;

-- name: FindWorkflowInstanceByID :one
SELECT id, tenant_id, workflow_id, workflow_version_id, correlation_key, status, version, current_state_instance_id, started_at, completed_at, expires_at, created_at, updated_at
FROM workflow_instances
WHERE id = $1 AND tenant_id = $2
LIMIT 1;

-- name: ListWorkflowInstancesByTenant :many
SELECT id, tenant_id, workflow_id, workflow_version_id, correlation_key, status, version, current_state_instance_id, started_at, completed_at, expires_at, created_at, updated_at
FROM workflow_instances
WHERE tenant_id = $1
ORDER BY created_at DESC;

-- name: UpdateWorkflowInstanceStatus :one
UPDATE workflow_instances
SET status = $4, version = version + 1, updated_at = NOW()
WHERE id = $1 AND tenant_id = $2 AND version = $3
RETURNING id, tenant_id, workflow_id, workflow_version_id, correlation_key, status, version, current_state_instance_id, started_at, completed_at, expires_at, created_at, updated_at;

-- name: IncrementWorkflowInstanceVersion :one
UPDATE workflow_instances
SET version = version + 1, updated_at = NOW()
WHERE id = $1 AND tenant_id = $2 AND version = $3
RETURNING id, tenant_id, workflow_id, workflow_version_id, correlation_key, status, version, current_state_instance_id, started_at, completed_at, expires_at, created_at, updated_at;

-- name: CreateStateInstance :one
INSERT INTO state_instances (tenant_id, workflow_instance_id, workflow_version_id, state_key, state_id, status)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, tenant_id, workflow_instance_id, workflow_version_id, state_key, state_id, status, version, retry_count, entered_at, expires_at, exited_at, created_at, updated_at;

-- name: FindStateInstanceByID :one
SELECT id, tenant_id, workflow_instance_id, workflow_version_id, state_key, state_id, status, version, retry_count, entered_at, expires_at, exited_at, created_at, updated_at
FROM state_instances
WHERE id = $1 AND tenant_id = $2
LIMIT 1;

-- name: UpdateStateInstanceStatus :one
UPDATE state_instances
SET status = $4, version = version + 1, updated_at = NOW()
WHERE id = $1 AND tenant_id = $2 AND version = $3
RETURNING id, tenant_id, workflow_instance_id, workflow_version_id, state_key, state_id, status, version, retry_count, entered_at, expires_at, exited_at, created_at, updated_at;

-- name: UpdateStateInstanceRetry :one
UPDATE state_instances
SET retry_count = retry_count + 1, version = version + 1, updated_at = NOW()
WHERE id = $1 AND tenant_id = $2 AND version = $3
RETURNING id, tenant_id, workflow_instance_id, workflow_version_id, state_key, state_id, status, version, retry_count, entered_at, expires_at, exited_at, created_at, updated_at;

-- name: ExitStateInstance :one
UPDATE state_instances
SET status = $4, exited_at = NOW(), version = version + 1, updated_at = NOW()
WHERE id = $1 AND tenant_id = $2 AND version = $3
RETURNING id, tenant_id, workflow_instance_id, workflow_version_id, state_key, state_id, status, version, retry_count, entered_at, expires_at, exited_at, created_at, updated_at;

-- name: SetCurrentStateInstance :exec
UPDATE workflow_instances
SET current_state_instance_id = $3, updated_at = NOW()
WHERE id = $1 AND tenant_id = $2;
