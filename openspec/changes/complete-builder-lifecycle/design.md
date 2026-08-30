## Context

See `proposal.md` for motivation. The prior Builder API introduced workflow roots,
optimistic versioning, immutable published snapshots, and React Query hooks, but it
stores only root metadata during a draft save. `draft-bridge.ts` therefore keeps the
graph and workflow id in localStorage; `publish()` exists in the Zustand store but
is not wired to the toolbar. The database already stores full published snapshots
as JSONB and workflow versions have stable node/transition ids suitable for a
structural comparison.

## Goals / Non-Goals

**Goals:**
- Make PostgreSQL the durable source of truth for an editable graph draft.
- Preserve optimistic concurrency and immutable published snapshots.
- Deliver an operator workflow for save, publish, history, and comparison without
  moving domain data-fetching into presentational components.

**Non-Goals:**
- Synchronizing open editors in real time, merging concurrent edits, or restoring
  an old version into the draft.
- Replacing the existing builder's local undo/redo history with server history.
- Expanding the scope to simulation, runtime inspection, or authorization UI.

## Decisions

### D1 — Store one mutable `draft_definition` JSONB on `workflows`

Add a non-null JSONB draft column to the workflow root, backfilled with a safe empty
object for pre-existing rows. The domain workflow and workflow DTO carry the raw
definition. Create and optimistic update receive a complete definition; get returns
it. Publish reads this stored definition rather than accepting a client-supplied
snapshot.

This separates the mutable authoring head from `workflow_versions.definition`, which
remains append-only. It is simpler and safer than creating a version row on every
autosave, and avoids the browser being the only durable copy.

- Alternative: keep drafts only as unversioned browser data. Rejected because it
  cannot survive device changes and conflicts with the phase goal.
- Alternative: create a version for every autosave. Rejected because it pollutes
  publish history and confuses operational rollback semantics.

### D2 — One optimistic write updates metadata and graph atomically

Extend the repository's workflow update operation to update name, nullable
description, and `draft_definition` under `WHERE version = expected_version`, then
increment `version`. The operation returns the fully updated root. API `PATCH`
requires the full definition and the current version. Existing roots with a missing
definition are readable as an empty draft only for migration compatibility, but
cannot be published until a valid graph has been saved.

The first State Builder save creates a workflow with its graph; each later autosave
uses the returned id/version. An HTTP 409 leaves the current React Flow state intact,
marks it conflicted, and asks the operator to reload. It never retries by force.

### D3 — Publish snapshots the persisted draft, with server validation

`POST /api/workflows/:id/publish` accepts only the expected workflow version. Within
the existing repository transaction it reads/uses the root's draft definition,
validates its JSON envelope and graph invariants, inserts version `current + 1`,
marks it current, and bumps the root version. The root remains editable after
publication; its draft head can diverge from the latest published snapshot before
the next publish.

The Go validation service is authoritative and returns structured rule violations
(rule code, message, optional node/transition id). The frontend's existing validator
is used for immediate feedback and a fast publish guard, but the API remains the
final authority. Validation rules must at least cover a non-empty graph, exactly one
START node, valid transition endpoints, and a terminal path; shared test fixtures
keep the two implementations aligned.

- Alternative: receive the graph in the publish body. Rejected because an unsaved
  or malicious client could publish a graph the operator never persisted, breaking
  the server-draft guarantee.

### D4 — Version APIs are read-only and compute a semantic graph diff

Add `GET /api/workflows/:id/versions/:versionNo` and
`GET /api/workflows/:id/versions/compare?baseVersion=&targetVersion=`. Both verify
tenant, project, workflow ownership, and distinct requested versions. The compare
service parses the two immutable definitions, maps nodes and transitions by stable
`id`, and emits deterministic added/removed/changed collections. Changed items carry
only top-level field names; guards/policies are represented as a changed field rather
than a noisy raw JSON patch.

This gives a domain-readable diff in the UI and keeps all clients consistent.

- Alternative: calculate a raw JSON patch in the browser. Rejected because it
  duplicates comparison logic, exposes unstable array-order differences, and makes
  the API incomplete for other clients.

### D5 — Route-driven load and a route-level lifecycle surface

The Builder accepts an optional workflow id in its route/query state. For an id it
hydrates from `useGetWorkflow`; without one it starts a new draft and replaces the
URL after first create. The store remains responsible for graph editing and debounced
save orchestration, while a route-level client container owns React Query mutations
and passes callbacks/status to the canvas. Presentational toolbar/history/diff
components receive data and handlers as props and never import axios or hooks.

The toolbar gains Publish and Versions actions. The history panel is available only
after the workflow has an id; it uses list/detail/compare hooks and shows a safe empty
state. Publish serializes with autosave so it cannot race a pending debounced write.

### D6 — Local browser data is a non-authoritative one-time migration only

Remove normal load/save use of `draft-bridge.ts` and PGlite. On first open without a
server workflow id, the implementation can offer to import a legacy local draft into
a new server workflow. It must never auto-overwrite a loaded server draft and must
clear the legacy keys only after a successful user-confirmed import.

## Risks / Trade-offs

- [Large JSON drafts increase row size/write volume] → Keep JSONB limited to the
  workflow DSL, debounce saves, and retain immutable versions only on publish.
- [Frontend/backend validation can drift] → Reuse fixtures in both test suites and
  have the backend enforce the minimum publish invariants.
- [Autosave and publish can race] → Flush/cancel the debounce and serialize both
  operations through one in-flight lifecycle state.
- [Legacy local draft could mask server state] → Treat it only as explicit migration
  input, never as hydration authority.

## Migration Plan

1. Add the additive draft-definition migration, queries, generated sqlc code, and
   repository/service tests before deploying the new UI.
2. Deploy API support first; existing workflows remain readable and can be resaved
   to establish a valid draft body.
3. Deploy the Builder UI, which uses API drafts for all new work and exposes the
   one-time legacy import where applicable.
4. Rollback the UI independently if required; no rollback deletes the new column or
   immutable snapshots. A database rollback is safe only before dependent application
   code is rolled back and must not remove data without an approved migration plan.
