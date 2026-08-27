-- name: FindUserByEmail :one
SELECT id, email, password_hash, name, role, status, photo, created_at, updated_at
FROM users
WHERE email = $1
LIMIT 1;

-- name: FindUserByID :one
SELECT id, email, password_hash, name, role, status, photo, created_at, updated_at
FROM users
WHERE id = $1
LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (email, password_hash, name, role, status)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, email, password_hash, name, role, status, photo, created_at, updated_at;

-- name: CreateAuthSession :one
INSERT INTO auth_sessions (user_id, token_hash, expires_at)
VALUES ($1, $2, $3)
RETURNING id, user_id, token_hash, expires_at, created_at;

-- name: FindSessionByTokenHash :one
SELECT id, user_id, token_hash, expires_at, created_at
FROM auth_sessions
WHERE token_hash = $1
  AND expires_at > NOW()
LIMIT 1;

-- name: DeleteSessionByTokenHash :exec
DELETE FROM auth_sessions
WHERE token_hash = $1;

-- name: DeleteSessionsByUserID :exec
DELETE FROM auth_sessions
WHERE user_id = $1;
