## 1. Go module rename

- [ ] 1.1 Update `go.work` to reference new module paths
- [ ] 1.2 `apps/api/go.mod`: `go mod edit -module github.com/irosadie/open-state/api`
- [ ] 1.3 `apps/worker/go.mod`: `go mod edit -module github.com/irosadie/open-state/worker`
- [ ] 1.4 `packages/go-shared/go.mod`: `go mod edit -module github.com/irosadie/open-state/go-shared`
- [ ] 1.5 Rewrite all Go imports `github.com/vibecoding-starter/*` → `github.com/irosadie/open-state/*`
- [ ] 1.6 Update `replace` directives (go-shared)
- [ ] 1.7 Run `go mod tidy` in all Go apps
- [ ] 1.8 Verify: `go build ./...`, `go vet ./...`, `go test ./...`

## 2. Node packages rename

- [ ] 2.1 Root `package.json` name → `openstate`
- [ ] 2.2 `packages/types|utils|schemas/package.json` name → `@openstate/*`
- [ ] 2.3 `apps/web/package.json` name + `@vibecoding-starter/*` deps → `@openstate/*`
- [ ] 2.4 `tsconfig.base.json` / `apps/web/tsconfig.json` path aliases
- [ ] 2.5 Update all `@vibecoding-starter/*` imports in frontend code
- [ ] 2.6 Update root `turbo.json` / workspaces references
- [ ] 2.7 Verify: `bun install`, `bun run typecheck`, `bun run lint`

## 3. Display branding

- [ ] 3.1 `apps/web/app/layout.tsx` metadata title/description → OpenState
- [ ] 3.2 `apps/api` system info (AppInfo name/version) → OpenState
- [ ] 3.3 Root package.json `description` → OpenState

## 4. Infrastructure naming

- [ ] 4.1 `docker-compose.yml` container names → `openstate-*`
- [ ] 4.2 Database name `vibecoding_starter` → `openstate` (compose + env)
- [ ] 4.3 `apps/api/.env.example`, `apps/worker/.env.example`, `apps/web/.env.example` updated
- [ ] 4.4 Verify stack: `docker compose up -d`, migrations run

## 5. Documentation & records

- [ ] 5.1 `docs/OPERATION.md` references updated
- [ ] 5.2 `docs/openapi.json`, `docs/openapi/base.json` updated
- [ ] 5.3 README (if any leftover references) updated
- [ ] 5.4 `openspec/changes/migrate-backend-to-golang/*` module references corrected

## 6. Final quality gate

- [ ] 6.1 `bun run check` (lint + typecheck + test + build) passes
- [ ] 6.2 `go build ./...`, `go vet ./...`, `go test ./...` passes (api + worker)
- [ ] 6.3 No remaining `vibecoding-starter|vibecoding` references (grep)
- [ ] 6.4 Smoke test: server starts, health check OK
