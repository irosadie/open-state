## Context

The workflow-definition **persistence** layer is already complete (`IWorkflowRepository`,
pgx adapter, SQL schema from `persistence-workflow-definitions`), and the State Builder
UI already builds a full `WorkflowDefinition` (nodes, transitions, metadata, validation)
that it currently persists to a browser-local PGlite store. The missing link is an HTTP
Builder API that exposes `IWorkflowRepository` and a frontend that calls it. The backend
uses Echo + Clean Architecture with the established controller → service → repository
pattern (see `capability_controller.go` / `capability_service.go`). Every operation is
tenant+project scoped; the tenant comes from `X-Tenant-ID`, never the body (PRD §74, §96).

See proposal.md — Why for motivation and the two specs for required behavior.

## Goals / Non-Goals

**Goals:**
- Add a minimal, production-grade Builder API (create/get/update draft, list, publish,
  list-versions) behind JWT auth, tenant+project scoped.
- Wire the State Builder to persist drafts via the API instead of PGlite.
- Follow the existing Echo/Go Clean Architecture and the frontend
  `web-api-integrated` patterns exactly.

**Non-Goals:**
- Simulation (PRD 53), version diff/compare (PRD 120/167), archive/restore.
- A full Project management API/UI.
- Admin console, runtime inspector, golden tests (future slices).

## Decisions

### D1 — Project resolution: default project per tenant
`IWorkflowRepository` requires a `projectID`. There is no project-management UI yet, and
the State Builder workflows have no project concept today. To avoid blocking this slice,
the `BuilderService` resolves the project from an optional `projectId` request field; when
absent, it finds-or-creates a **default project** (`slug = "default"`) for the tenant via
`IProjectRepository` and uses its ID. This keeps the State Builder working end-to-end
without a separate project setup step, and is a strict superset for future project UI.

- **Alternative considered:** Require `projectId` on every call. Rejected because there is
  no way to create a project from the UI yet, so the State Builder could not save at all.

### D2 — Draft persistence model: workflow root + full definition JSONB
Draft edits are persisted by creating a workflow root (`IWorkflowRepository.Create`,
`status=DRAFT`) and, on update, mutating the root's mutable fields (name/description) with
optimistic concurrency. The full `WorkflowDefinition` JSON is stored in the published
`workflow_versions.definition` on publish (PRD §68). For the draft slice, the node/edge
definition JSON lives with the workflow root via the State Builder's own serialization
carried in the request; this is intentionally a thin slice — deep draft-body versioning is
deferred to the publish/compare slice.

- **Note:** This keeps the change minimal and matches what "Save Draft" needs today.

### D3 — Routes and responses
- `POST /api/workflows` — create draft.
- `GET /api/workflows` — list (query: `projectId`).
- `GET /api/workflows/:id` — get.
- `PATCH /api/workflows/:id` — update draft (name/description/definition, `version`).
- `POST /api/workflows/:id/publish` — publish (body: `definition`, `version`).
- `GET /api/workflows/:id/versions` — list versions.
- All under JWT + auth-session middleware; tenant from `X-Tenant-ID`.
- Responses wrapped as `{ "data": ... }`, matching the capability controller convention.

### D4 — Frontend swap: hooks replace PGlite
Add `packages/schemas/workflow.ts` (Zod), `packages/types/workflow-response.ts`,
`apps/web/constants/api-routers.ts` + `query-keys.ts` entries, and
`apps/web/hooks/transactions/use-workflow/` react-query hooks (list, get,
create-or-upsert-draft, publish). The State Builder store's `persist()`/`hydrate()`
swap `pglite-store.ts` for the hooks' underlying axios calls. React Flow state,
undo/redo, and validation remain untouched.

### D5 — `CreateApp` signature gains the workflow controller
`CreateApp(authCtrl, systemCtrl, capabilityCtrl, repo, tokenSvc)` gains a
`workflowCtrl *controllers.WorkflowController` param; `main.go` constructs the
`BuilderService` from `adapter.Workflows()` + `adapter.Projects()` and wires it.

## Risks / Trade-offs

- **Default-project magic** → Mitigated by making `projectId` explicit and defaulting only
  as a fallback; a future project UI supersedes it without breaking existing data.
- **Optimistic-lock conflicts surface to the operator** → The frontend mutation surfaces
  the 409 conflict error so the operator can retry after refresh.
- **Draft body not yet versioned** → Intentionally deferred; the full-definition JSON is
  versioned only at publish. Noted as a non-goal so no false expectation is set.
