-- name: ListRoutableIntents :many
SELECT i.id,
       i.tenant_id,
       i.project_id,
       i.workflow_id,
       i.intent_key,
       i.name,
       i.description,
       i.examples,
       i.created_at,
       i.updated_at,
       w.slug AS workflow_slug
FROM intents i
JOIN projects p
  ON p.id = i.project_id
 AND p.tenant_id = i.tenant_id
JOIN workflows w
  ON w.id = i.workflow_id
 AND w.tenant_id = i.tenant_id
 AND w.project_id = i.project_id
WHERE i.tenant_id = $1
  AND i.project_id = $2
  AND p.status = 'ACTIVE'
  AND w.status = 'PUBLISHED'
ORDER BY i.created_at ASC, i.intent_key ASC;

-- name: FindRoutableIntent :one
SELECT i.id,
       i.tenant_id,
       i.project_id,
       i.workflow_id,
       i.intent_key,
       i.name,
       i.description,
       i.examples,
       i.created_at,
       i.updated_at,
       w.slug AS workflow_slug
FROM intents i
JOIN projects p
  ON p.id = i.project_id
 AND p.tenant_id = i.tenant_id
JOIN workflows w
  ON w.id = i.workflow_id
 AND w.tenant_id = i.tenant_id
 AND w.project_id = i.project_id
WHERE i.tenant_id = $1
  AND i.project_id = $2
  AND i.intent_key = $3
  AND p.status = 'ACTIVE'
  AND w.status = 'PUBLISHED'
LIMIT 1;

-- name: UpsertIntent :one
INSERT INTO intents (
  tenant_id,
  project_id,
  workflow_id,
  intent_key,
  name,
  description,
  examples
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (tenant_id, project_id, intent_key)
DO UPDATE SET
  workflow_id = EXCLUDED.workflow_id,
  name = EXCLUDED.name,
  description = EXCLUDED.description,
  examples = EXCLUDED.examples,
  updated_at = NOW()
RETURNING id,
          tenant_id,
          project_id,
          workflow_id,
          intent_key,
          name,
          description,
          examples,
          created_at,
          updated_at;
