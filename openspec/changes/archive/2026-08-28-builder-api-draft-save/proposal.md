## Why

Epic **#5 (State Builder produksi + Admin console)** task #1 requires the State Builder's
"Save Draft" to persist to the backend engine instead of the browser-local PGlite
(PRD 146, PRD 128). Today the workflow-definition **persistence** layer is complete
(`IWorkflowRepository` + pgx adapter + SQL schema, from `persistence-workflow-definitions`),
but there is **no HTTP surface** exposing it — the State Builder UI can only save drafts
into an embedded browser database. This change delivers the **Builder API** (the HTTP
contract behind PRD 146: create/get/update/validate/publish/list-versions/compare/archive)
as the first production-grade slice of the Frontend epic, and wires the State Builder to it.

## What Changes

- **NEW** — `BuilderService` in `apps/api/internal/application/services/` orchestrating
  workflow-definition operations over `IWorkflowRepository` (tenant+project scoped, PRD §4, §96).
- **NEW** — request/response DTOs in `apps/api/internal/application/dtos/workflow.go`
  for create/get/update draft + list + publish + list-versions (PRD 146).
- **NEW** — `WorkflowController` in `apps/api/internal/interfaces/http/controllers/`
  parsing requests, calling the service, formatting `{ "data": ... }` responses.
- **NEW** — route registration in `apps/api/internal/interfaces/http/routes/routes.go`
  under `/api/workflows` behind JWT + auth-session middleware.
- **NEW** — wiring in `apps/api/cmd/server/main.go` composition root (service → controller → `CreateApp`).
- **NEW** — frontend shared contract: `packages/schemas/workflow.ts` (Zod), `packages/types/workflow-response.ts`, routes in `apps/web/constants/api-routers.ts`, query keys in `apps/web/constants/query-keys.ts`.
- **NEW** — react-query hooks in `apps/web/hooks/transactions/use-workflow/` (list, get, create/upsert draft, publish).
- **MODIFIED** — State Builder persistence layer: swap the PGlite draft store in
  `apps/web/components/state-builder/utils/pglite-store.ts` for the new API hooks
  (create/get/update draft), keeping the in-memory React Flow state intact.
- **NEW** — OpenAPI docs under `docs/openapi/` for the new workflow endpoints.

## Capabilities

### New Capabilities

- `backend/builder-api`: the HTTP Builder API (PRD 146) exposing workflow-definition
  draft CRUD, publish-to-immutable-version, list, and version listing behind auth, all
  tenant+project scoped via `IWorkflowRepository`.
- `web/state-builder-api`: the frontend integration of the State Builder with the
  Builder API — shared Zod schemas, response types, constants, react-query hooks, and the
  swap from PGlite to API-backed draft persistence.

### Modified Capabilities

- `backend/persistence/workflow-definitions`: the persistence contract already defines the
  `IWorkflowRepository` interface; the Builder API is the first consumer exposing it over
  HTTP. No persistence requirements change — only new consumption. No delta spec required.

## Impact

- **`apps/api/internal/application/services/builder_service.go`** — new service.
- **`apps/api/internal/application/dtos/workflow.go`** — new DTOs.
- **`apps/api/internal/interfaces/http/controllers/workflow_controller.go`** — new controller.
- **`apps/api/internal/interfaces/http/routes/routes.go`** — register workflow routes.
- **`apps/api/internal/interfaces/http/create_app.go`** — add controller param.
- **`apps/api/cmd/server/main.go`** — composition root wiring.
- **`packages/schemas/workflow.ts`**, **`packages/types/workflow-response.ts`** — new shared contract.
- **`apps/web/constants/api-routers.ts`**, **`query-keys.ts`** — new routes/keys.
- **`apps/web/hooks/transactions/use-workflow/`** — new react-query hooks.
- **`apps/web/components/state-builder/`** — swap PGlite → API persistence.
- **`docs/openapi/`** — new workflow endpoint docs.

## Non-Goals

- Full PRD 146 simulation (PRD 53), compare-versions diff (PRD 120/167), and archive/restore —
  future slices of epic #5, kept minimal here.
- Runtime instance admin (PRD 142, 147) and Admin console — separate slices.
- Changing the workflow-definition **schema/persistence** — already delivered.
- Frontend golden-conversation tests (PRD 125), simulation mode, RBAC UI — future slices.

## Dependencies

- `backend/persistence/workflow-definitions` (IWorkflowRepository + pgx adapter) — already archived.
- Epic #3 (schema) and #5 (frontend epic).
- Auth/JWT middleware (already present in `apps/api`).
