## 1. Backend — Builder API (Skill: api-feature)

- [x] 1.1 Read `.agents/guides/ARCHITECTURE.md`, `api-dto.md`, `api-service.md`, `api-controller.md`, `api-route.md`; review `capability_controller.go`/`capability_service.go` as the reference pattern.
- [x] 1.2 Create `apps/api/internal/application/dtos/workflow.go` — request/response DTOs: CreateWorkflowRequest, UpdateWorkflowRequest, PublishWorkflowRequest, WorkflowDTO, WorkflowVersionDTO, WorkflowListDTO.
- [x] 1.3 Create `apps/api/internal/application/services/builder_service.go` — `BuilderService` over `IWorkflowRepository` + `IProjectRepository`: CreateDraft, Get, List, UpdateDraft (optimistic), Publish, ListVersions; default-project resolution (design D1).
- [x] 1.4 Create `apps/api/internal/application/services/builder_service_test.go` — unit tests for create/get/update (optimistic conflict), publish, and default-project resolution (no DB; use a fake repository).
- [x] 1.5 Create `apps/api/internal/interfaces/http/controllers/workflow_controller.go` — `WorkflowController` parsing requests, calling the service, returning `{ "data": ... }` (PRD §74).
- [x] 1.6 Register workflow routes in `apps/api/internal/interfaces/http/routes/routes.go` under `/api/workflows` behind JWT + auth-session middleware (PRD §74).
- [x] 1.7 Update `apps/api/internal/interfaces/http/create_app.go` to accept the `WorkflowController`.
- [x] 1.8 Wire `BuilderService` + `WorkflowController` in `apps/api/cmd/server/main.go` composition root (from `adapter.Workflows()` + `adapter.Projects()`).
- [x] 1.9 Run `go build ./...` and `go test ./...` in `apps/api`; fix failures.

## 2. Backend — OpenAPI (Skill: docs-openapi)

- [x] 2.1 Write `docs/openapi` split docs for the new `/api/workflows` endpoints (create/get/update/list/publish/versions).
- [x] 2.2 Run the openapi generation/validation script and confirm no drift.

## 3. Frontend — Shared contracts (Skill: web-api-integrated)

- [x] 3.1 Read `.agents/guides/shared-schema.md`, `web-type.md`, `web-constant.md`, `web-hook.md`; review `apps/web/hooks/transactions/use-capability/` as the reference.
- [x] 3.2 Create `packages/schemas/workflow.ts` — Zod schemas for create/update/publish workflow drafts; typed status constants.
- [x] 3.3 Create `packages/types/workflow-response.ts` — Workflow, WorkflowVersion, WorkflowList response types.
- [x] 3.4 Add workflow routes to `apps/web/constants/api-routers.ts` (create/list/get/update/publish/versions with `:id` path variables).
- [x] 3.5 Add workflow query keys to `apps/web/constants/query-keys.ts`.
- [x] 3.6 Add `packages/schemas/workflow.test.ts` — schema validation unit tests.

## 4. Frontend — Hooks (Skill: web-api-integrated)

- [x] 4.1 Create `apps/web/hooks/transactions/use-workflow/use-list-workflows.ts` (useQuery).
- [x] 4.2 Create `use-get-workflow.ts` (useQuery by id).
- [x] 4.3 Create `use-create-workflow.ts` (useMutation → create draft; invalidate list).
- [x] 4.4 Create `use-update-workflow.ts` (useMutation → update draft with optimistic `version`; invalidate list).
- [x] 4.5 Create `use-publish-workflow.ts` (useMutation → publish; invalidate list + versions).
- [x] 4.6 Create `use-list-workflow-versions.ts` (useQuery by id).
- [x] 4.7 Export all from `apps/web/hooks/transactions/use-workflow/index.ts`.

## 5. Frontend — State Builder integration

- [x] 5.1 Read `apps/web/components/state-builder/state-builder.store.ts` and `utils/pglite-store.ts` to identify the persist/hydrate swap points.
- [x] 5.2 Replace the PGlite draft persistence in the State Builder with the workflow API hooks (create-or-upsert draft on persist; load by id on hydrate); keep React Flow state, undo/redo, validation intact.
- [x] 5.3 Surface save status (saving/saved/error) in the State Builder toolbar via the existing store fields.
- [x] 5.4 Run `bun run test` and `bun run build` in `apps/web`; add/adjust smoke tests if the repo has them.
- [x] 5.5 Biome clean: no `any`, no `console.*`, no unused imports/vars, 2-space indent, double quotes, no semicolons (read root `biome.json`).

## 6. Quality gate (Skills: web-code-review, api-code-review)

- [x] 6.1 `go build ./...` + `go test ./...` in `apps/api`.
- [x] 6.2 `bun run test` + `bun run build` in `apps/web`.
- [x] 6.3 No JSX calling axios/fetch directly (all data via hooks).
- [x] 6.4 No business logic in controller; no DB access in use case/service beyond repository.
- [x] 6.5 All files end with a newline (EOF).
