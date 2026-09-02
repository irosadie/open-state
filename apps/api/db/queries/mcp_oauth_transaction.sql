-- name: CreateMCPAuthorizationTransaction :one
INSERT INTO mcp_oauth_transactions (tenant_id, project_id, connection_id, state_hash, verifier_reference, redirect_uri, expires_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, tenant_id, project_id, connection_id, state_hash, verifier_reference,
  redirect_uri, expires_at, status, created_at, updated_at;

-- name: FindPendingMCPAuthorizationTransaction :one
SELECT id, tenant_id, project_id, connection_id, state_hash, verifier_reference,
  redirect_uri, expires_at, status, created_at, updated_at
FROM mcp_oauth_transactions
WHERE tenant_id = $1 AND project_id = $2 AND connection_id = $3
  AND state_hash = $4 AND status = 'pending' AND expires_at > NOW()
LIMIT 1;

-- name: ConsumeMCPAuthorizationTransaction :execrows
UPDATE mcp_oauth_transactions
SET status = 'consumed', updated_at = NOW()
WHERE id = $1 AND tenant_id = $2 AND project_id = $3 AND status = 'pending';
