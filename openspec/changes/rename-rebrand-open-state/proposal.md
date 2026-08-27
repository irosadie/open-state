## Why

The repository was cloned from `vibecoding-starter-go`, but the project it now hosts is **OpenState** — an enterprise conversation state orchestration platform. The GitHub repo is already named `irosadie/open-state`, yet the internal identity is still `vibecoding-starter`: package names, Go module paths, application title/description, container names, and documentation all reference the old starter brand.

This inconsistency hurts:
- **Discoverability & branding** — the product is "OpenState", not "vibecoding-starter".
- **Import hygiene** — Go module `github.com/vibecoding-starter/*` is misleading and non-portable.
- **Open-source trust** — a rebranded repo with old starter names looks unfinished.

This change renames/rebrands the entire repository to **OpenState**, keeping it consistent end-to-end (display name, packages, modules, containers, docs).

## What Changes

- **Display branding** — app title, description, root `package.json` name → `openstate` / `@openstate/*`.
- **Go module paths** — `github.com/vibecoding-starter/{api,worker,go-shared}` → `github.com/irosadie/open-state/{api,worker,go-shared}`.
- **All Go imports** updated to the new module path; `go.work` & `go.mod` files updated.
- **Frontend packages** — `@vibecoding-starter/types|utils|schemas` → `@openstate/types|utils|schemas`.
- **Docker/compose** — container names & image refs → `openstate-*`.
- **Docs & env examples** — `docs/OPERATION.md`, `docs/openapi*`, `.env.example`, README references.
- **Existing OpenSpec change** — `openspec/changes/migrate-backend-to-golang/*` references to `vibecoding-starter` updated (historical record kept but corrected).
- **DB name** — `vibecoding_starter` database in compose/env → `openstate`.

## Capabilities

### Modified Capabilities

- `monorepo/identity`: root `package.json` name, workspaces, turbo config reference the new `@openstate/*` scope.
- `backend/go-module`: Go module paths across `apps/api`, `apps/worker`, `packages/go-shared` renamed; all imports updated.
- `frontend/identity`: TypeScript package names & app metadata (title/description) reflect OpenState.
- `infra/containers`: docker-compose service/container names and image tags use `openstate-*`.
- `docs/identity`: documentation and env examples reference OpenState.

## Impact

- **`apps/api`, `apps/worker`, `packages/go-shared`** — `go.mod` module path changed; every internal import rewritten (mechanical, ~dozens of files).
- **`apps/web`** — package.json name, tsconfig paths (`@vibecoding-starter/*` → `@openstate/*`), app metadata.
- **`packages/*`** — package names & import refs.
- **Root** — `package.json`, `docker-compose.yml`, `.env*`, README, docs.
- **CI / quality** — must re-run: `bun run typecheck`, `bun run lint`, `go build ./...`, `go vet ./...`, tests.
- **No runtime behavior change** — purely cosmetic/naming. No workflow logic touched.

## Non-Goals

- No change to product features, workflow engine, or PRD content.
- No GitHub repository rename (already `irosadie/open-state`).
- No migration of existing git history.
