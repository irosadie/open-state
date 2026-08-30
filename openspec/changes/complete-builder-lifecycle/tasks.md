## 1. Persistence and domain contract

- [x] 1.1 Read the `db-sqlc-schema` skill, architecture guide, and existing workflow migration/query patterns before changing persistence.
- [x] 1.2 Add an additive goose migration that introduces a durable `draft_definition` JSONB field on workflow roots, safely backfills existing rows, and defines a reversible down migration.
- [x] 1.3 Extend workflow sqlc queries for create, read, and one optimistic update of metadata plus draft definition; regenerate sqlc code and update the pgx mapping.
- [x] 1.4 Extend the workflow entity and repository contract with the draft definition and atomic optimistic-update operation; remove or narrow metadata-only update use as appropriate.
- [x] 1.5 Add repository tests for draft round-trip, scoped reads, optimistic conflicts, and preservation of existing published version definitions.

## 2. Builder API and workflow validation

- [x] 2.1 Read the `api-feature` and `docs-openapi` skills plus the existing Builder service/controller tests and API conventions.
- [x] 2.2 Extend create, get, and update DTOs/responses so a complete draft graph is accepted, persisted, and returned; keep tenant and project resolution behavior intact.
- [x] 2.3 Implement the authoritative Go draft-graph validator with structured validation issues and test fixtures covering valid graphs, missing/duplicate START, invalid transition endpoints, and missing terminal path.
- [x] 2.4 Change publish to validate and snapshot the persisted server draft in the existing atomic repository transaction, using only the request's expected optimistic version.
- [x] 2.5 Add read-only API/service operations for a specific published version and for comparison of two distinct versions of one workflow.
- [x] 2.6 Implement deterministic node/transition diff calculation keyed by stable ids, including added, removed, changed, and changed-field results.
- [x] 2.7 Register and test the expanded authenticated/RBAC-protected routes, including scope, malformed compare inputs, 404, validation, and 409 responses.
- [x] 2.8 Update split OpenAPI documents and regenerate/validate the checked-in OpenAPI artifact for all changed and new workflow endpoints.

## 3. Shared frontend API integration

- [x] 3.1 Read the `web-api-integrated` skill, shared-schema/type/hook guides, and existing workflow hook patterns.
- [x] 3.2 Extend Zod request schemas, response types, API-router constants, and query keys for durable drafts, version detail, validation issues, and graph diff results; add schema tests.
- [x] 3.3 Update existing create/get/update/publish hooks to use the new contract and precise invalidation; add hooks for version detail and comparison.
- [x] 3.4 Add hook tests for request shaping, cache invalidation, and error propagation without direct HTTP from JSX.

## 4. State Builder lifecycle UI

- [x] 4.1 Read the `web-slicing` skill and inspect the State Builder's store/canvas/toolbar composition before UI changes.
- [x] 4.2 Add a route-level Builder lifecycle container that owns workflow query/mutation hooks and passes lifecycle state/handlers into presentational Builder components.
- [x] 4.3 Replace normal PGlite/localStorage draft hydration and persistence with server draft create/update/load, maintaining the in-memory graph, undo/redo, and debounced autosave behavior.
- [x] 4.4 Add route identifier handling so a persisted workflow can be reopened directly and first save updates the Builder URL without losing unsaved in-memory work.
- [x] 4.5 Implement save-state and conflict UX: saving, saved, failed, and reload-required conflict states must preserve the local graph and avoid forced retry.
- [x] 4.6 Wire a Publish button into the toolbar; serialize it with pending autosave, locally block invalid graphs, surface server validation/conflict failures, and display the successfully published version.
- [x] 4.7 Add a version-history panel with newest-first metadata/current marker, an explicit empty state, and selection controls for two distinct published versions.
- [x] 4.8 Add a read-only graph-diff view that renders node and transition additions, removals, and changed fields for the selected base/target pair.
- [x] 4.9 Implement an explicit, one-time legacy-local-draft import only if it cannot overwrite an existing server draft; remove obsolete bridge/PGlite use after a successful migration path is verified.

## 5. Verification and quality gate

- [x] 5.1 Add backend service/controller tests for complete draft round-trip, valid/invalid publish, immutable history, version detail, deterministic diff, tenant/project isolation, and optimistic conflicts.
- [x] 5.2 Add State Builder component/integration tests for API hydration, autosave state, publish success/failure, history empty/populated states, and diff rendering.
- [x] 5.3 Run `go test ./...` and `go build ./...` in `apps/api`; fix all failures.
- [x] 5.4 Run `bun run test` and `bun run build` in `apps/web`; fix all failures and Biome violations.
- [x] 5.5 Run `openspec validate` in strict mode for `complete-builder-lifecycle`, then review the implementation against every requirement and task before starting OpenSpec verification/archive.
