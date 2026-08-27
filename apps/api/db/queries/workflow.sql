-- name: CreateProject :one
INSERT INTO projects (tenant_id, name, slug, status)
VALUES ($1, $2, $3, $4)
RETURNING id, tenant_id, name, slug, status, created_at, updated_at;

-- name: FindProjectByID :one
SELECT id, tenant_id, name, slug, status, created_at, updated_at
FROM projects
WHERE id = $1 AND tenant_id = $2
LIMIT 1;

-- name: FindProjectBySlug :one
SELECT id, tenant_id, name, slug, status, created_at, updated_at
FROM projects
WHERE slug = $1 AND tenant_id = $2
LIMIT 1;

-- name: ListProjectsByTenant :many
SELECT id, tenant_id, name, slug, status, created_at, updated_at
FROM projects
WHERE tenant_id = $1
ORDER BY created_at DESC;

-- name: CreateWorkflow :one
INSERT INTO workflows (tenant_id, project_id, slug, name, description, status)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, tenant_id, project_id, slug, name, description, status, current_version, version, created_at, updated_at;

-- name: FindWorkflowByID :one
SELECT id, tenant_id, project_id, slug, name, description, status, current_version, version, created_at, updated_at
FROM workflows
WHERE id = $1 AND tenant_id = $2 AND project_id = $3
LIMIT 1;

-- name: FindWorkflowBySlug :one
SELECT id, tenant_id, project_id, slug, name, description, status, current_version, version, created_at, updated_at
FROM workflows
WHERE slug = $1 AND tenant_id = $2 AND project_id = $3
LIMIT 1;

-- name: ListWorkflowsByTenant :many
SELECT id, tenant_id, project_id, slug, name, description, status, current_version, version, created_at, updated_at
FROM workflows
WHERE tenant_id = $1 AND project_id = $2
ORDER BY created_at DESC;

-- name: UpdateWorkflowStatus :one
UPDATE workflows
SET status = $4, updated_at = NOW()
WHERE id = $1 AND tenant_id = $2 AND project_id = $3 AND version = $5
RETURNING id, tenant_id, project_id, slug, name, description, status, current_version, version, created_at, updated_at;

-- name: UpdateWorkflowVersion :one
UPDATE workflows
SET version = version + 1, current_version = $4, updated_at = NOW()
WHERE id = $1 AND tenant_id = $2 AND project_id = $3 AND version = $5
RETURNING id, tenant_id, project_id, slug, name, description, status, current_version, version, created_at, updated_at;

-- name: CreateWorkflowVersion :one
INSERT INTO workflow_versions (workflow_id, tenant_id, project_id, version_no, definition, status, is_current)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, workflow_id, tenant_id, project_id, version_no, definition, status, is_current, created_at, updated_at;

-- name: FindWorkflowVersionByNumber :one
SELECT id, workflow_id, tenant_id, project_id, version_no, definition, status, is_current, created_at, updated_at
FROM workflow_versions
WHERE workflow_id = $1 AND version_no = $2 AND tenant_id = $3 AND project_id = $4
LIMIT 1;

-- name: FindCurrentWorkflowVersion :one
SELECT id, workflow_id, tenant_id, project_id, version_no, definition, status, is_current, created_at, updated_at
FROM workflow_versions
WHERE workflow_id = $1 AND tenant_id = $2 AND project_id = $3 AND is_current = TRUE
LIMIT 1;

-- name: ListWorkflowVersions :many
SELECT id, workflow_id, tenant_id, project_id, version_no, definition, status, is_current, created_at, updated_at
FROM workflow_versions
WHERE workflow_id = $1 AND tenant_id = $2 AND project_id = $3
ORDER BY version_no DESC;

-- name: SetCurrentWorkflowVersion :exec
UPDATE workflow_versions
SET is_current = (version_no = $2), updated_at = NOW()
WHERE workflow_id = $1 AND tenant_id = $3 AND project_id = $4;

-- name: CreateState :one
INSERT INTO states (workflow_version_id, key, kind, name, description, instructions, required_context, capabilities, policy, is_terminal, position)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING id, workflow_version_id, key, kind, name, description, instructions, required_context, capabilities, policy, is_terminal, position, created_at, updated_at;

-- name: ListStatesByVersion :many
SELECT s.id, s.workflow_version_id, s.key, s.kind, s.name, s.description, s.instructions,
       s.required_context, s.capabilities, s.policy, s.is_terminal, s.position, s.created_at, s.updated_at
FROM states s
JOIN workflow_versions wv ON wv.id = s.workflow_version_id
WHERE s.workflow_version_id = $1 AND wv.tenant_id = $2 AND wv.project_id = $3
ORDER BY s.created_at ASC;

-- name: CreateTransition :one
INSERT INTO transitions (workflow_version_id, key, source_state_id, target_state_id, event, priority, is_active)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, workflow_version_id, key, source_state_id, target_state_id, event, priority, is_active, created_at, updated_at;

-- name: ListTransitionsByVersion :many
SELECT t.id, t.workflow_version_id, t.key, t.source_state_id, t.target_state_id, t.event, t.priority, t.is_active, t.created_at, t.updated_at
FROM transitions t
JOIN workflow_versions wv ON wv.id = t.workflow_version_id
WHERE t.workflow_version_id = $1 AND wv.tenant_id = $2 AND wv.project_id = $3
ORDER BY t.priority ASC;

-- name: CreateTransitionGuard :one
INSERT INTO transition_guards (transition_id, workflow_version_id, logic, conditions)
VALUES ($1, $2, $3, $4)
RETURNING id, transition_id, workflow_version_id, logic, conditions, created_at, updated_at;

-- name: ListGuardsByTransition :many
SELECT g.id, g.transition_id, g.workflow_version_id, g.logic, g.conditions, g.created_at, g.updated_at
FROM transition_guards g
JOIN transitions t ON t.id = g.transition_id
JOIN workflow_versions wv ON wv.id = t.workflow_version_id
WHERE g.transition_id = $1 AND wv.tenant_id = $2 AND wv.project_id = $3
ORDER BY g.created_at ASC;
