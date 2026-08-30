## Why

Epic #5 Phase 1 is only partially complete: the Builder API can create workflow
roots and publish versions, but the editable graph still lives in browser storage
and the UI does not expose publishing, version history, or comparison. This means
an operator can lose a draft when changing browser/device and cannot safely review
what changes between published workflow versions.

## What Changes

- Persist the complete editable `WorkflowDefinition` graph as the server-side draft
  head, tenant and project scoped, with optimistic concurrency.
- Return the persisted draft definition when loading a workflow, and require draft
  saves and publishes to use the same authoritative server-side graph.
- Add version-detail and version-comparison API operations. Version snapshots stay
  immutable; the server returns canonical structured graph changes instead of a
  browser-specific textual diff.
- Replace the State Builder localStorage bridge with API-backed hydration and
  autosave, using a workflow identifier in the builder route so a saved workflow
  can be reopened directly.
- Add a Publish action that saves the current draft first, blocks invalid graphs,
  communicates request/conflict errors, and creates a new immutable version.
- Add a State Builder version-history panel with published-version metadata and a
  selected pair diff for added, removed, and changed nodes and transitions.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `backend/persistence/workflow-definitions`: workflow roots gain a durable draft
  definition and repository operations required to read and update it.
- `backend/builder-api`: draft create/read/update and version APIs expose the full
  graph, a specific immutable version, and a canonical version comparison.
- `web/state-builder-api`: the Builder uses the API as its draft source of truth
  and provides publish, history, and graph-diff operator controls.

## Impact

- Database migration and sqlc query regeneration in `apps/api/db/`, workflow domain
  entities, repository contract, pgx adapter, Builder service/controller/routes,
  OpenAPI contract, and backend tests.
- Shared workflow schemas/types, API route and query-key constants, React Query
  hooks, and State Builder route/store/components/tests in `apps/web/` and
  `packages/`.
- `localStorage` is no longer an authoritative draft store. The implementation may
  make a one-time best-effort migration only when no server draft exists; it must
  not overwrite a server draft.

## Non-goals

- Running workflows in a simulation sandbox (Phase 2).
- Runtime instance inspection/debugging, Admin Console expansion, RBAC UI
  enforcement, and frontend golden/E2E suites (Phases 3–6).
- Restoring or mutating a published version, archive/restore lifecycle controls,
  collaborative live editing, or general-purpose JSON diffing outside workflow
  graphs.
