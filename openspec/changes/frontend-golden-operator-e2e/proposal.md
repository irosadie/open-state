# Change: Add frontend golden journeys for Builder and operator flows

## Why

The platform has deterministic engine golden tests and backend MCP E2E coverage, but it has no browser-level regression suite for the workflows users actually operate. State Builder lifecycle, permission-aware routing, Admin Console runtime controls, and Runtime Inspector can each pass isolated tests while their combined user journey regresses.

## What Changes

- Add a dedicated browser E2E runner and isolated application test environment separate from the existing Vitest/jsdom suite.
- Add version-controlled golden journey manifests and synthetic tenant fixtures for State Builder and operator runtime flows.
- Cover the Builder journey from authenticated Editor access through draft save/reload, valid publish, version history, and graph diff, including invalid or stale-save behavior.
- Cover the Operator journey from authorized landing and runtime discovery through instance detail, sanitized debug evidence, lifecycle command confirmation, persisted result, and audit verification.
- Assert permission-aware route/action behavior in those journeys, including denied routes or controls that must not fetch protected data.
- Run the browser suite in CI with useful but secret-safe failure diagnostics; block all real LLM, RAG, MCP, and external provider traffic.

## Capabilities

### New Capabilities

- `quality/frontend-golden-e2e`: Provides deterministic, fixture-driven browser golden journeys for State Builder and tenant-scoped operator runtime operations.

### Modified Capabilities

None.

## Impact

- Affected frontend areas: browser-test configuration/scripts, stable UI test affordances, State Builder, Admin Console, Runtime Inspector, and Auth/RBAC UI route/action behavior.
- Affected test infrastructure: ephemeral Postgres/Redis, non-production fixture seed/verification utilities, CI browser installation and artifact handling.
- Depends on the active `complete-builder-lifecycle`, `runtime-inspector-debug`, `admin-console-management`, and `auth-rbac-ui-guards` changes to publish their routes, permissions, and contracts.
- Does not replace existing engine golden conversations or backend MCP E2E tests; it validates the browser-to-platform integration layer.

## Non-goals

- Screenshot-based visual regression as the primary golden signal, load testing, or comprehensive cross-browser compatibility certification.
- Running against production data, real identity providers, or any external LLM, RAG, MCP, observability, or capability provider.
- Re-testing every unit/API scenario already covered by narrower suites.
