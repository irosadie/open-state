-- name: AppendEvent :one
INSERT INTO events (tenant_id, event_id, type, source, aggregate_id, workflow_instance_id, timestamp, payload, correlation_id, causation_id, idempotency_key)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING id, tenant_id, event_id, type, source, aggregate_id, workflow_instance_id, sequence, timestamp, payload, correlation_id, causation_id, idempotency_key, created_at;

-- name: FindEventByID :one
SELECT id, tenant_id, event_id, type, source, aggregate_id, workflow_instance_id, sequence, timestamp, payload, correlation_id, causation_id, idempotency_key, created_at
FROM events
WHERE id = $1 AND tenant_id = $2
LIMIT 1;

-- name: ListEventsByInstance :many
SELECT id, tenant_id, event_id, type, source, aggregate_id, workflow_instance_id, sequence, timestamp, payload, correlation_id, causation_id, idempotency_key, created_at
FROM events
WHERE workflow_instance_id = $1 AND tenant_id = $2
ORDER BY sequence;

-- name: ListEventsByTenant :many
SELECT id, tenant_id, event_id, type, source, aggregate_id, workflow_instance_id, sequence, timestamp, payload, correlation_id, causation_id, idempotency_key, created_at
FROM events
WHERE tenant_id = $1
ORDER BY sequence;

-- name: ListEventsFiltered :many
SELECT id, tenant_id, event_id, type, source, aggregate_id, workflow_instance_id, sequence, timestamp, payload, correlation_id, causation_id, idempotency_key, created_at
FROM events
WHERE tenant_id = @tenant_id
  AND (sqlc.narg('workflow_instance_id')::uuid IS NULL OR workflow_instance_id = sqlc.narg('workflow_instance_id'))
  AND (sqlc.narg('type')::text IS NULL OR type = sqlc.narg('type'))
  AND (sqlc.narg('source')::text IS NULL OR source = sqlc.narg('source'))
  AND (sqlc.narg('correlation_id')::text IS NULL OR correlation_id = sqlc.narg('correlation_id'))
ORDER BY sequence DESC
LIMIT @page_size OFFSET @page_offset;

-- name: CountEventsFiltered :one
SELECT COUNT(*)
FROM events
WHERE tenant_id = @tenant_id
  AND (sqlc.narg('workflow_instance_id')::uuid IS NULL OR workflow_instance_id = sqlc.narg('workflow_instance_id'))
  AND (sqlc.narg('type')::text IS NULL OR type = sqlc.narg('type'))
  AND (sqlc.narg('source')::text IS NULL OR source = sqlc.narg('source'))
  AND (sqlc.narg('correlation_id')::text IS NULL OR correlation_id = sqlc.narg('correlation_id'));

-- name: InsertInboxEvent :one
INSERT INTO event_inbox (tenant_id, idempotency_key, event_type, source, payload)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, tenant_id, idempotency_key, event_type, source, payload, status, attempt_count, received_at, processed_at, created_at, updated_at;

-- name: ClaimInboxEvents :many
UPDATE event_inbox
SET status = 'PROCESSING', attempt_count = attempt_count + 1, updated_at = NOW()
WHERE id IN (
  SELECT sub.id FROM event_inbox AS sub
  WHERE sub.tenant_id = $1 AND sub.status = 'RECEIVED'
  ORDER BY sub.received_at
  LIMIT $2
)
RETURNING id, tenant_id, idempotency_key, event_type, source, payload, status, attempt_count, received_at, processed_at, created_at, updated_at;

-- name: MarkInboxProcessed :one
UPDATE event_inbox
SET status = 'PROCESSED', processed_at = NOW(), updated_at = NOW()
WHERE id = $1 AND tenant_id = $2
RETURNING id, tenant_id, idempotency_key, event_type, source, payload, status, attempt_count, received_at, processed_at, created_at, updated_at;

-- name: InsertOutboxEvent :one
INSERT INTO event_outbox (tenant_id, event_id, payload, topic)
VALUES ($1, $2, $3, $4)
RETURNING id, tenant_id, event_id, payload, topic, status, attempt_count, published_at, created_at, updated_at;

-- name: ClaimOutboxEvents :many
UPDATE event_outbox
SET status = 'PENDING', attempt_count = attempt_count + 1, updated_at = NOW()
WHERE id IN (
  SELECT sub.id FROM event_outbox AS sub
  WHERE sub.tenant_id = $1 AND sub.status = 'PENDING'
  ORDER BY sub.created_at
  LIMIT $2
)
RETURNING id, tenant_id, event_id, payload, topic, status, attempt_count, published_at, created_at, updated_at;

-- name: MarkOutboxPublished :one
UPDATE event_outbox
SET status = 'PUBLISHED', published_at = NOW(), updated_at = NOW()
WHERE id = $1 AND tenant_id = $2
RETURNING id, tenant_id, event_id, payload, topic, status, attempt_count, published_at, created_at, updated_at;

-- name: UpsertIdempotencyRecord :one
INSERT INTO idempotency_records (tenant_id, idempotency_key, scope, result_status, payload)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (tenant_id, idempotency_key) DO UPDATE
SET scope = EXCLUDED.scope, result_status = EXCLUDED.result_status, payload = EXCLUDED.payload, updated_at = NOW()
RETURNING id, tenant_id, idempotency_key, scope, result_status, payload, created_at, updated_at;

-- name: FindIdempotencyRecord :one
SELECT id, tenant_id, idempotency_key, scope, result_status, payload, created_at, updated_at
FROM idempotency_records
WHERE tenant_id = $1 AND idempotency_key = $2
LIMIT 1;
