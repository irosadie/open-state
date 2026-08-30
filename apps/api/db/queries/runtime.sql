-- Runtime Inspector read projections. Every query is tenant-scoped.

-- name: FindRuntimeWorkflow :one
SELECT id, tenant_id, project_id, slug, name, description, status, current_version, version, created_at, updated_at, draft_definition
FROM workflows
WHERE id = $1 AND tenant_id = $2
LIMIT 1;

-- name: FindRuntimeWorkflowVersion :one
SELECT id, workflow_id, tenant_id, project_id, version_no, definition, status, is_current, created_at, updated_at
FROM workflow_versions
WHERE id = $1 AND tenant_id = $2
LIMIT 1;

-- name: ListRuntimeStatesByVersion :many
SELECT s.id, s.workflow_version_id, s.key, s.kind, s.name, s.description, s.instructions,
       s.required_context, s.capabilities, s.policy, s.is_terminal, s.position, s.created_at, s.updated_at
FROM states s
JOIN workflow_versions wv ON wv.id = s.workflow_version_id
WHERE s.workflow_version_id = $1 AND wv.tenant_id = $2
ORDER BY s.created_at ASC, s.id ASC;

-- name: ListStateInstancesByWorkflowInstance :many
SELECT id, tenant_id, workflow_instance_id, workflow_version_id, state_key, state_id, status,
       version, retry_count, entered_at, expires_at, exited_at, created_at, updated_at
FROM state_instances
WHERE workflow_instance_id = $1 AND tenant_id = $2
ORDER BY entered_at ASC, created_at ASC, id ASC;

-- name: AppendRuntimeTraceEntry :one
INSERT INTO runtime_trace_entries (
  tenant_id, workflow_instance_id, turn_id, stage, source, status, occurred_at,
  correlation_id, duration_ms, reason_code, error_code, provider_alias,
  provider_reference, summary, attributes
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
RETURNING id, tenant_id, workflow_instance_id, turn_id, sequence, stage, source, status,
          occurred_at, correlation_id, duration_ms, reason_code, error_code,
          provider_alias, provider_reference, summary, attributes;

-- name: ListRuntimeTraceByInstance :many
SELECT id, tenant_id, workflow_instance_id, turn_id, sequence, stage, source, status,
       occurred_at, correlation_id, duration_ms, reason_code, error_code,
       provider_alias, provider_reference, summary, attributes
FROM runtime_trace_entries
WHERE tenant_id = $1 AND workflow_instance_id = $2
ORDER BY sequence ASC, occurred_at ASC, id ASC;

-- name: ListRuntimeTraceByTurn :many
SELECT id, tenant_id, workflow_instance_id, turn_id, sequence, stage, source, status,
       occurred_at, correlation_id, duration_ms, reason_code, error_code,
       provider_alias, provider_reference, summary, attributes
FROM runtime_trace_entries
WHERE tenant_id = $1 AND workflow_instance_id = $2 AND turn_id = $3
ORDER BY sequence ASC, occurred_at ASC, id ASC;
