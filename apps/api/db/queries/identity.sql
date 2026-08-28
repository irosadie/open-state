-- External identity queries (PRD 79). Links an OIDC provider subject to a local
-- user and supports auto-provisioning on first login.

-- name: FindIdentityByProviderSubject :one
SELECT id, user_id, provider, subject_id, auto_provisioned, created_at, updated_at
FROM user_identities
WHERE provider = $1 AND subject_id = $2
LIMIT 1;

-- name: CreateIdentity :one
INSERT INTO user_identities (user_id, provider, subject_id, auto_provisioned)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, provider, subject_id, auto_provisioned, created_at, updated_at;

-- name: ListIdentitiesByUser :many
SELECT id, user_id, provider, subject_id, auto_provisioned, created_at, updated_at
FROM user_identities
WHERE user_id = $1
ORDER BY created_at DESC;
