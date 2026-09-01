## 1. Go module rename

- [x] 1.1 Update `go.work` to reference new module paths
- [x] 1.2 `apps/api/go.mod`: `go mod edit -module github.com/irosadie/open-state/api`
- [x] 1.3 `apps/worker/go.mod`: `go mod edit -module github.com/irosadie/open-state/worker`
- [x] 1.4 `packages/go-shared/go.mod`: `go mod edit -module github.com/irosadie/open-state/go-shared`
- [x] 1.5 Verify all Go imports use `github.com/irosadie/open-state/*`
- [x] 1.6 Update `replace` directives (go-shared)
- [x] 1.7 Run `go mod tidy` in all Go apps
- [x] 1.8 Verify: `go build ./...`, `go vet ./...`, `go test ./...`

## 2. Node packages rename

- [x] 2.1 Root `package.json` name → `openstate`
- [x] 2.2 `packages/types|utils|schemas/package.json` name → `@openstate/*`
- [x] 2.3 `apps/web/package.json` name and workspace dependencies use `@openstate/*`
- [x] 2.4 `tsconfig.base.json` / `apps/web/tsconfig.json` path aliases
- [x] 2.5 Verify all frontend imports use `@openstate/*`
- [x] 2.6 Update root `turbo.json` / workspaces references
- [x] 2.7 Verify: `bun install`, `bun run typecheck`, `bun run lint`

## 3. Display branding

- [x] 3.1 `apps/web/app/layout.tsx` metadata title/description → OpenState
- [x] 3.2 `apps/api` system info (AppInfo name/version) → OpenState
- [x] 3.3 Root package.json `description` → OpenState

## 4. Infrastructure naming

- [x] 4.1 `docker-compose.yml` container names → `openstate-*`
- [x] 4.2 Database name is `openstate` (compose + env)
- [x] 4.3 `apps/api/.env.example`, `apps/worker/.env.example`, `apps/web/.env.example` updated
- [x] 4.4 Verify stack: `docker compose up -d`, migrations run

## 5. Documentation & records

- [x] 5.1 `docs/OPERATION.md` references updated
- [x] 5.2 `docs/openapi.json`, `docs/openapi/base.json` updated
- [x] 5.3 README (if any leftover references) updated
- [x] 5.4 `openspec/changes/migrate-backend-to-golang/*` module references corrected

## 6. Final quality gate

- [x] 6.1 `bun run check` (lint + typecheck + test + build) passes
- [x] 6.2 `go build ./...`, `go vet ./...`, `go test ./...` passes (api + worker)
- [x] 6.3 No remaining legacy starter identity references (grep)
- [x] 6.4 Smoke test: server starts, health check OK
