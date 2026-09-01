-- name: UpsertCapabilityExecutionEvidence :one
INSERT INTO capability_execution_evidence (
  tenant_id, project_id, workflow_instance_id, state_id, capability_id,
  capability_name, provider_server, provider_tool, correlation_id,
  idempotency_key, status, result, error
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
ON CONFLICT (tenant_id, workflow_instance_id, state_id, capability_id, idempotency_key)
DO UPDATE SET
  provider_server = EXCLUDED.provider_server,
  provider_tool = EXCLUDED.provider_tool,
  correlation_id = EXCLUDED.correlation_id,
  status = EXCLUDED.status,
  result = EXCLUDED.result,
  error = EXCLUDED.error,
  updated_at = NOW()
RETURNING id, tenant_id, project_id, workflow_instance_id, state_id, capability_id,
  capability_name, provider_server, provider_tool, correlation_id, idempotency_key,
  status, result, error, created_at, updated_at;

-- name: FindCapabilityEvidenceByIdempotency :one
SELECT id, tenant_id, project_id, workflow_instance_id, state_id, capability_id,
  capability_name, provider_server, provider_tool, correlation_id, idempotency_key,
  status, result, error, created_at, updated_at
FROM capability_execution_evidence
WHERE tenant_id = $1 AND project_id = $2 AND workflow_instance_id = $3
  AND state_id = $4 AND capability_id = $5 AND idempotency_key = $6
LIMIT 1;

-- name: ListCapabilityEvidenceByState :many
SELECT id, tenant_id, project_id, workflow_instance_id, state_id, capability_id,
  capability_name, provider_server, provider_tool, correlation_id, idempotency_key,
  status, result, error, created_at, updated_at
FROM capability_execution_evidence
WHERE tenant_id = $1 AND project_id = $2 AND workflow_instance_id = $3 AND state_id = $4
ORDER BY created_at ASC;

-- name: ListCapabilityEvidenceByInstanceState :many
SELECT id, tenant_id, project_id, workflow_instance_id, state_id, capability_id,
  capability_name, provider_server, provider_tool, correlation_id, idempotency_key,
  status, result, error, created_at, updated_at
FROM capability_execution_evidence
WHERE tenant_id = $1 AND workflow_instance_id = $2 AND state_id = $3
ORDER BY created_at ASC;
