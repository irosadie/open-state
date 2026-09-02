-- name: CreateMCPConnection :one
INSERT INTO mcp_connections (
  tenant_id, project_id, name, alias, transport, endpoint, stdio_profile,
  stdio_args, auth_type, credential_reference, credential_status, status,
  oauth_authorization_endpoint, oauth_token_endpoint, oauth_client_id,
  oauth_client_secret_reference, oauth_scopes, oauth_redirect_uri, timeout_ms, max_concurrency,
  rate_limit_per_second, rate_limit_burst, retry_max, circuit_failure_threshold,
  circuit_recovery_seconds, created_by, updated_by
)
VALUES (
  $1, $2, $3, $4, $5, NULLIF($6, ''), NULLIF($7, ''), $8, $9,
  NULLIF($10, ''), $11, $12, NULLIF($13, ''), NULLIF($14, ''), NULLIF($15, ''),
  NULLIF($16, ''), $17, NULLIF($18, ''), $19, $20, $21, $22, $23, $24, $25, $26, $26
)
RETURNING id, tenant_id, project_id, name, alias, transport, endpoint,
  stdio_profile, stdio_args, auth_type, credential_reference, credential_status,
  status, oauth_authorization_endpoint, oauth_token_endpoint, oauth_client_id,
  oauth_client_secret_reference, oauth_scopes, oauth_redirect_uri, oauth_access_token_reference,
  oauth_refresh_token_reference, oauth_expires_at, oauth_status, health_status,
  health_reason, last_success_at, consecutive_failures, circuit_opened_at,
  timeout_ms, max_concurrency, rate_limit_per_second, rate_limit_burst,
  retry_max, circuit_failure_threshold, circuit_recovery_seconds,
  last_test_status, last_test_error_code, last_tested_at, created_by,
  updated_by, created_at, updated_at;

-- name: FindMCPConnectionByID :one
SELECT id, tenant_id, project_id, name, alias, transport, endpoint,
  stdio_profile, stdio_args, auth_type, credential_reference, credential_status,
  status, oauth_authorization_endpoint, oauth_token_endpoint, oauth_client_id,
  oauth_client_secret_reference, oauth_scopes, oauth_redirect_uri, oauth_access_token_reference,
  oauth_refresh_token_reference, oauth_expires_at, oauth_status, health_status,
  health_reason, last_success_at, consecutive_failures, circuit_opened_at,
  timeout_ms, max_concurrency, rate_limit_per_second, rate_limit_burst,
  retry_max, circuit_failure_threshold, circuit_recovery_seconds,
  last_test_status, last_test_error_code, last_tested_at, created_by,
  updated_by, created_at, updated_at
FROM mcp_connections
WHERE id = $1 AND tenant_id = $2 AND project_id = $3
LIMIT 1;

-- name: ListMCPConnectionsByProject :many
SELECT id, tenant_id, project_id, name, alias, transport, endpoint,
  stdio_profile, stdio_args, auth_type, credential_reference, credential_status,
  status, oauth_authorization_endpoint, oauth_token_endpoint, oauth_client_id,
  oauth_client_secret_reference, oauth_scopes, oauth_redirect_uri, oauth_access_token_reference,
  oauth_refresh_token_reference, oauth_expires_at, oauth_status, health_status,
  health_reason, last_success_at, consecutive_failures, circuit_opened_at,
  timeout_ms, max_concurrency, rate_limit_per_second, rate_limit_burst,
  retry_max, circuit_failure_threshold, circuit_recovery_seconds,
  last_test_status, last_test_error_code, last_tested_at, created_by,
  updated_by, created_at, updated_at
FROM mcp_connections
WHERE tenant_id = $1 AND project_id = $2
ORDER BY created_at DESC, alias ASC;

-- name: UpdateMCPConnection :one
UPDATE mcp_connections
SET name = $4,
    alias = $5,
    transport = $6,
    endpoint = NULLIF($7, ''),
    stdio_profile = NULLIF($8, ''),
    stdio_args = $9,
    auth_type = $10,
    credential_reference = CASE WHEN $11 IS NULL THEN NULL WHEN $11 = '' THEN credential_reference ELSE $11 END,
    credential_status = $12,
    oauth_authorization_endpoint = NULLIF($13, ''),
    oauth_token_endpoint = NULLIF($14, ''),
    oauth_client_id = NULLIF($15, ''),
    oauth_client_secret_reference = NULLIF($16, ''),
    oauth_scopes = $17,
    oauth_redirect_uri = NULLIF($18, ''),
    timeout_ms = $19,
    max_concurrency = $20,
    rate_limit_per_second = $21,
    rate_limit_burst = $22,
    retry_max = $23,
    circuit_failure_threshold = $24,
    circuit_recovery_seconds = $25,
    updated_by = $26,
    updated_at = NOW()
WHERE id = $1 AND tenant_id = $2 AND project_id = $3
RETURNING id, tenant_id, project_id, name, alias, transport, endpoint,
  stdio_profile, stdio_args, auth_type, credential_reference, credential_status,
  status, oauth_authorization_endpoint, oauth_token_endpoint, oauth_client_id,
  oauth_client_secret_reference, oauth_scopes, oauth_redirect_uri, oauth_access_token_reference,
  oauth_refresh_token_reference, oauth_expires_at, oauth_status, health_status,
  health_reason, last_success_at, consecutive_failures, circuit_opened_at,
  timeout_ms, max_concurrency, rate_limit_per_second, rate_limit_burst,
  retry_max, circuit_failure_threshold, circuit_recovery_seconds,
  last_test_status, last_test_error_code, last_tested_at, created_by,
  updated_by, created_at, updated_at;

-- name: DeleteMCPConnection :exec
DELETE FROM mcp_connections
WHERE id = $1 AND tenant_id = $2 AND project_id = $3;

-- name: UpdateMCPConnectionStatus :one
UPDATE mcp_connections
SET status = $4,
    last_test_status = CASE WHEN $4 = 'disabled' THEN 'disabled' ELSE last_test_status END,
    updated_by = $5,
    updated_at = NOW()
WHERE id = $1 AND tenant_id = $2 AND project_id = $3
RETURNING id, tenant_id, project_id, name, alias, transport, endpoint,
  stdio_profile, stdio_args, auth_type, credential_reference, credential_status,
  status, oauth_authorization_endpoint, oauth_token_endpoint, oauth_client_id,
  oauth_client_secret_reference, oauth_scopes, oauth_redirect_uri, oauth_access_token_reference,
  oauth_refresh_token_reference, oauth_expires_at, oauth_status, health_status,
  health_reason, last_success_at, consecutive_failures, circuit_opened_at,
  timeout_ms, max_concurrency, rate_limit_per_second, rate_limit_burst,
  retry_max, circuit_failure_threshold, circuit_recovery_seconds,
  last_test_status, last_test_error_code, last_tested_at, created_by,
  updated_by, created_at, updated_at;

-- name: RecordMCPConnectionTest :one
UPDATE mcp_connections
SET last_test_status = $4,
    last_test_error_code = NULLIF($5, ''),
    last_tested_at = NOW(),
    updated_by = $6,
    updated_at = NOW()
WHERE id = $1 AND tenant_id = $2 AND project_id = $3
RETURNING id, tenant_id, project_id, name, alias, transport, endpoint,
  stdio_profile, stdio_args, auth_type, credential_reference, credential_status,
  status, oauth_authorization_endpoint, oauth_token_endpoint, oauth_client_id,
  oauth_client_secret_reference, oauth_scopes, oauth_redirect_uri, oauth_access_token_reference,
  oauth_refresh_token_reference, oauth_expires_at, oauth_status, health_status,
  health_reason, last_success_at, consecutive_failures, circuit_opened_at,
  timeout_ms, max_concurrency, rate_limit_per_second, rate_limit_burst,
  retry_max, circuit_failure_threshold, circuit_recovery_seconds,
  last_test_status, last_test_error_code, last_tested_at, created_by,
  updated_by, created_at, updated_at;

-- name: UpdateMCPConnectionOAuth :one
UPDATE mcp_connections
SET credential_reference = $4,
    oauth_access_token_reference = $4,
    oauth_refresh_token_reference = $5,
    credential_status = $6,
    oauth_status = $7,
    oauth_expires_at = $8,
    updated_by = $9,
    updated_at = NOW()
WHERE id = $1 AND tenant_id = $2 AND project_id = $3
RETURNING id, tenant_id, project_id, name, alias, transport, endpoint,
  stdio_profile, stdio_args, auth_type, credential_reference, credential_status,
  status, oauth_authorization_endpoint, oauth_token_endpoint, oauth_client_id,
  oauth_client_secret_reference, oauth_scopes, oauth_redirect_uri, oauth_access_token_reference,
  oauth_refresh_token_reference, oauth_expires_at, oauth_status, health_status,
  health_reason, last_success_at, consecutive_failures, circuit_opened_at,
  timeout_ms, max_concurrency, rate_limit_per_second, rate_limit_burst,
  retry_max, circuit_failure_threshold, circuit_recovery_seconds,
  last_test_status, last_test_error_code, last_tested_at, created_by,
  updated_by, created_at, updated_at;

-- name: UpdateMCPConnectionCredential :one
UPDATE mcp_connections
SET credential_reference = $4,
    credential_status = $5,
    updated_by = $6,
    updated_at = NOW()
WHERE id = $1 AND tenant_id = $2 AND project_id = $3
RETURNING id, tenant_id, project_id, name, alias, transport, endpoint,
  stdio_profile, stdio_args, auth_type, credential_reference, credential_status,
  status, oauth_authorization_endpoint, oauth_token_endpoint, oauth_client_id,
  oauth_client_secret_reference, oauth_scopes, oauth_redirect_uri, oauth_access_token_reference,
  oauth_refresh_token_reference, oauth_expires_at, oauth_status, health_status,
  health_reason, last_success_at, consecutive_failures, circuit_opened_at,
  timeout_ms, max_concurrency, rate_limit_per_second, rate_limit_burst,
  retry_max, circuit_failure_threshold, circuit_recovery_seconds,
  last_test_status, last_test_error_code, last_tested_at, created_by,
  updated_by, created_at, updated_at;

-- name: DisconnectMCPConnectionOAuth :one
UPDATE mcp_connections
SET credential_reference = NULL,
    oauth_access_token_reference = NULL,
    oauth_refresh_token_reference = NULL,
    oauth_expires_at = NULL,
    oauth_status = 'disconnected',
    credential_status = 'action_required',
    updated_by = $4,
    updated_at = NOW()
WHERE id = $1 AND tenant_id = $2 AND project_id = $3
RETURNING id, tenant_id, project_id, name, alias, transport, endpoint,
  stdio_profile, stdio_args, auth_type, credential_reference, credential_status,
  status, oauth_authorization_endpoint, oauth_token_endpoint, oauth_client_id,
  oauth_client_secret_reference, oauth_scopes, oauth_redirect_uri, oauth_access_token_reference,
  oauth_refresh_token_reference, oauth_expires_at, oauth_status, health_status,
  health_reason, last_success_at, consecutive_failures, circuit_opened_at,
  timeout_ms, max_concurrency, rate_limit_per_second, rate_limit_burst,
  retry_max, circuit_failure_threshold, circuit_recovery_seconds,
  last_test_status, last_test_error_code, last_tested_at, created_by,
  updated_by, created_at, updated_at;

-- name: RecordMCPConnectionHealth :one
UPDATE mcp_connections
SET health_status = $4,
    health_reason = NULLIF($5, ''),
    last_success_at = COALESCE($6, last_success_at),
    consecutive_failures = $7,
    circuit_opened_at = $8,
    updated_by = $9,
    updated_at = NOW()
WHERE id = $1 AND tenant_id = $2 AND project_id = $3
RETURNING id, tenant_id, project_id, name, alias, transport, endpoint,
  stdio_profile, stdio_args, auth_type, credential_reference, credential_status,
  status, oauth_authorization_endpoint, oauth_token_endpoint, oauth_client_id,
  oauth_client_secret_reference, oauth_scopes, oauth_redirect_uri, oauth_access_token_reference,
  oauth_refresh_token_reference, oauth_expires_at, oauth_status, health_status,
  health_reason, last_success_at, consecutive_failures, circuit_opened_at,
  timeout_ms, max_concurrency, rate_limit_per_second, rate_limit_burst,
  retry_max, circuit_failure_threshold, circuit_recovery_seconds,
  last_test_status, last_test_error_code, last_tested_at, created_by,
  updated_by, created_at, updated_at;

-- name: ResetMCPConnectionHealth :one
UPDATE mcp_connections
SET health_status = 'unknown', health_reason = NULL, consecutive_failures = 0,
    circuit_opened_at = NULL, updated_by = $4, updated_at = NOW()
WHERE id = $1 AND tenant_id = $2 AND project_id = $3
RETURNING id, tenant_id, project_id, name, alias, transport, endpoint,
  stdio_profile, stdio_args, auth_type, credential_reference, credential_status,
  status, oauth_authorization_endpoint, oauth_token_endpoint, oauth_client_id,
  oauth_client_secret_reference, oauth_scopes, oauth_redirect_uri, oauth_access_token_reference,
  oauth_refresh_token_reference, oauth_expires_at, oauth_status, health_status,
  health_reason, last_success_at, consecutive_failures, circuit_opened_at,
  timeout_ms, max_concurrency, rate_limit_per_second, rate_limit_burst,
  retry_max, circuit_failure_threshold, circuit_recovery_seconds,
  last_test_status, last_test_error_code, last_tested_at, created_by,
  updated_by, created_at, updated_at;
