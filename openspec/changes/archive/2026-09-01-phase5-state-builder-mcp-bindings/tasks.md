## 1. Binding persistence and resolution

- [x] 1.1 Verify Phase 3–4 project connection and discovered-tool contracts, including active/removed/disabled statuses.
- [x] 1.2 Add a project capability MCP binding migration with foreign keys and same-project integrity constraints.
- [x] 1.3 Add sqlc queries and regenerate code for binding CRUD, binding health, and scoped provider-requirement resolution.
- [x] 1.4 Add domain entities/repositories/services that validate capability, project connection, and discovered tool ownership as one transaction.
- [x] 1.5 Preserve legacy provider alias/tool fields as compatibility metadata and define explicit missing-binding status.

## 2. Authoring and publishing validation

- [x] 2.1 Add capability/Builder API operations that list only eligible project connection/tool pairs.
- [x] 2.2 Validate binding creation/update against connection and tool health without accepting raw URLs or tool-name text.
- [x] 2.3 Add save and publish validation that rejects required missing, disabled, removed, or stale bindings with actionable errors.
- [x] 2.4 Update State MCP provider requirement resolution to prefer the project binding and return safe availability only.

## 3. State Builder experience

- [x] 3.1 Add shared binding schemas/types/constants and React Query hooks for the project MCP catalog and binding mutations.
- [x] 3.2 Replace free-text provider alias/tool inputs in the capability authoring form with project connection and enabled-tool selectors.
- [x] 3.3 Add binding-health indicators to State Builder editing, validation, preview, save, and publish feedback.
- [x] 3.4 Add empty-catalog guidance that routes operators to project MCP Connections without offering a raw target fallback.

## 4. Migration and verification

- [x] 4.1 Add deterministic seed/backfill handling only for exact project connection/tool matches; mark ambiguous rows missing.
- [x] 4.2 Add backend tests for scope isolation, invalid targets, health transitions, publish rejection, and safe State MCP projections.
- [x] 4.3 Add State Builder tests for filtered selectors, unavailable binding feedback, and no direct HTTP from components.
- [x] 4.4 Run migration/sqlc checks, Go tests/vet, web tests/typecheck/lint, MCP contract tests, and OpenSpec validation.
