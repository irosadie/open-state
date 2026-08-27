-- +goose Up
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TYPE user_role AS ENUM ('USER', 'ADMIN');
CREATE TYPE user_status AS ENUM ('ACTIVE', 'SUSPENDED');

CREATE TABLE users (
  id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  email         VARCHAR(255) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,
  name          VARCHAR(255) NOT NULL,
  role          user_role   NOT NULL DEFAULT 'USER',
  status        user_status NOT NULL DEFAULT 'ACTIVE',
  photo         VARCHAR(512),
  created_at    TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMP(6) NOT NULL DEFAULT NOW()
);

CREATE TABLE auth_sessions (
  id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash  VARCHAR(255) NOT NULL,
  expires_at  TIMESTAMP(6) NOT NULL,
  created_at  TIMESTAMP(6) NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS auth_sessions;
DROP TABLE IF EXISTS users;
DROP TYPE IF EXISTS user_status;
DROP TYPE IF EXISTS user_role;
