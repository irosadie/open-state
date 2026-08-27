---
name: db-sqlc-schema
description: Write or modify SQL migration files (goose) and sqlc query files, then regenerate Go code. Use for database model, relation, index, enum changes, or schema-to-domain-layer sync.
---

# Skill: DB sqlc Schema

## Context (Required)
- Folder scope + code samples: `references/context.md`
- Execution checklist: `templates/checklist.md`

Apply schema changes minimally and safely. Goose manages migrations, sqlc generates type-safe Go from SQL queries.

## Naming Conventions

### Tables
| Category | Table Prefix |
|---|---|
| Master / reference data | `master_` |
| Transaction / business data | `business_` |
| Membership data | `member_` |
| User / auth data | `users`, `auth_*` |
| System config data | `configurations` |

### Columns
- snake_case always
- Boolean fields — prefix `is_` or `has_`
- Enum-like fields — use `VARCHAR` + Go const, not PostgreSQL ENUM (prefer consistency)

### Required Standard Columns
```sql
id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
created_at TIMESTAMP(6) NOT NULL DEFAULT NOW(),
updated_at TIMESTAMP(6) NOT NULL DEFAULT NOW()
```

## Workflow

1. Write goose migration SQL in `apps/api/db/migrations/`
2. Write or update sqlc query in `apps/api/db/queries/`
3. Run: `sqlc generate` from `apps/api/`
4. Update domain entity + repository interface if schema changes affect them
5. Update `pgx_{domain}_repository.go` to use new generated query

## Goose Migration Format

```sql
-- +goose Up
CREATE TABLE master_categories (
  id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  name       VARCHAR(100) NOT NULL,
  is_active  BOOLEAN      NOT NULL DEFAULT TRUE,
  created_at TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP(6) NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS master_categories;
```

## sqlc Query Format

```sql
-- name: FindCategoryByID :one
SELECT id, name, is_active, created_at, updated_at
FROM master_categories
WHERE id = $1;

-- name: CreateCategory :one
INSERT INTO master_categories (name, is_active)
VALUES ($1, $2)
RETURNING id, name, is_active, created_at, updated_at;

-- name: ListCategories :many
SELECT id, name, is_active, created_at, updated_at
FROM master_categories
ORDER BY created_at DESC;
```

## Prohibitions

- **NEVER** edit files in `internal/infrastructure/db/` — they are sqlc-generated. Always regenerate.
- **NEVER** rename a table or column without a clear migration that handles the rename.
- **NEVER** drop an existing column without documenting the migration impact.
- **NEVER** use PostgreSQL ENUM type — use VARCHAR with Go typed constants instead.
- **NEVER** add a migration without a corresponding `-- +goose Down` block.

## Pre-Completion Checklist

- [ ] Migration file created with sequential filename (`000XX_description.sql`)
- [ ] `-- +goose Up` and `-- +goose Down` blocks present
- [ ] Standard columns present: id, created_at, updated_at
- [ ] sqlc query file created/updated
- [ ] `sqlc generate` run — no errors
- [ ] Domain entity updated if shape changed
- [ ] Repository interface updated if new queries added
- [ ] pgx repository implementation updated
- [ ] `go build ./...` passes
- [ ] All files end with a newline (EOF)
