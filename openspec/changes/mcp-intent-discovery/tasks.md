## 1. Intent persistence

- [x] 1.1 Add a goose migration for a tenant/project-scoped `intents` table with canonical key uniqueness, workflow foreign key, metadata, examples, and lookup indexes.
- [x] 1.2 Add sqlc queries for listing routable intents, finding one by canonical key, and idempotently upserting seed mappings; regenerate and verify the checked-in database code.

## 2. Domain and application catalog

- [x] 2.1 Add the intent domain entity and repository contract for scoped list/find operations.
- [x] 2.2 Implement the PostgreSQL intent repository and map query rows into the domain entity without weakening tenant/project scope.
- [x] 2.3 Refactor the intent application service to list catalog records and resolve only published workflow mappings by canonical key, including validation and not-found errors.
- [x] 2.4 Update dependency wiring and service tests for the new repository-backed intent contract.

## 3. MCP routing tools

- [x] 3.1 Extend the MCP intent projection to include natural-language examples and the stable canonical ID.
- [x] 3.2 Register the read-only `list_intents` tool with required tenant/project scope and return the documented `intents` response shape.
- [x] 3.3 Update `resolve_intent` to use canonical catalog keys, remove arbitrary workflow ID/slug fallback, and preserve the existing tool error format.
- [x] 3.4 Add MCP unit/e2e coverage for catalog discovery, `BOOKING_PADEL` selection metadata, scope isolation, unknown intents, and incomplete arguments.

## 4. Demo data and documentation

- [x] 4.1 Extend the demo seed with idempotent catalog records for `BOOKING_PADEL`, `ORDER_FOOD`, and `ORDER_DOCTOR`, including Indonesian-friendly example utterances.
- [x] 4.2 Document the routing flow: call `list_intents` → select the canonical `id` → call `resolve_intent` → continue with runtime/context tools.

## 5. Verification

- [x] 5.1 Run backend formatting, vet, unit/integration tests, and build checks against a migrated database.
- [x] 5.2 Start the MCP server with seeded data and verify live `tools/list`, `list_intents`, and `resolve_intent` calls, including `BOOKING_PADEL`.
