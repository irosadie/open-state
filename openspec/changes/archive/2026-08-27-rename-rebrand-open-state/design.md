## Overview

Mechanical rename of the repository identity from `vibecoding-starter` to
`openstate` / `OpenState`, with module path `github.com/irosadie/open-state/*`.

## Decisions

### D1. Go module path
`github.com/vibecoding-starter/{api,worker,go-shared}` →
`github.com/irosadie/open-state/{api,worker,go-shared}`

- `go.work` updated to reference new module paths.
- All `import "github.com/vibecoding-starter/..."` rewritten with
  `go mod edit -module` + sed for imports, or full `goimports` pass.
- Verification: `go build ./...`, `go vet ./...`, `go test ./...` in each app.

### D2. Node package scope
`@vibecoding-starter/{types,utils,schemas}` → `@openstate/{types,utils,schemas}`

- Root workspaces, each `packages/*/package.json`, `apps/web/package.json`.
- `tsconfig.base.json` / `apps/web/tsconfig.json` path aliases updated.
- Verification: `bun run typecheck`, `bun run lint`.

### D3. App display branding
- Root `package.json` `name` → `"openstate"`.
- `apps/web` metadata title/description → "OpenState".
- `apps/api` system info (name/version) → OpenState.

### D4. Infrastructure naming
- `docker-compose.yml` container names `vibecoding-starter-*` →
  `openstate-*`.
- Database name `vibecoding_starter` → `openstate` (compose + `.env.example`
  + `DATABASE_URL` defaults).
- `.env.example` files updated.

### D5. Documentation & records
- `docs/OPERATION.md`, `docs/openapi*.json`, README references updated.
- `openspec/changes/migrate-backend-to-golang/*` — historical references
  corrected to `github.com/irosadie/open-state/*` where they are still accurate
  (module path at time of writing).

## Risks / Notes
- **Go**: module path change requires updating every import; missed ones fail
  compile (safe — caught by `go build`).
- **Frontend**: path alias change caught by `tsc`/biome.
- **Historical OpenSpec change**: editing old records is cosmetic; accepted for
  consistency.

## Migration steps
1. Rename Go modules (`go mod edit -module`), update imports, run `go build`.
2. Rename Node packages + tsconfig aliases, run typecheck/lint.
3. Update branding (package.json names, app metadata).
4. Update docker-compose, env examples, DB name.
5. Update docs (OPERATION, openapi, README).
6. Full quality check: `bun run check` + `go build/vet/test`.
