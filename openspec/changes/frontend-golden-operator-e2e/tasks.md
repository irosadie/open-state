## 1. Browser E2E foundation

- [x] 1.1 Select and add the dedicated browser E2E runner, web workspace scripts, and isolated configuration without changing the existing Vitest/unit command.
- [x] 1.2 Define deterministic local web/API service orchestration with disposable Postgres and Redis plus health/readiness checks.
- [x] 1.3 Add a non-production fixture seed command and verification utility that use internal test/database access only and cannot be exposed through product HTTP routes.
- [x] 1.4 Add version-controlled golden journey manifest/types and stable UI test-affordance conventions for semantic checkpoints.
- [x] 1.5 Add runner smoke coverage proving browser startup, authenticated BFF request forwarding, fixture reset, and teardown work locally.

## 2. Deterministic fixtures and safety controls

- [x] 2.1 Seed fixed tenant-A Editor, Operator, Viewer, and sentinel tenant-B identities with the exact RBAC assignments needed by the golden journeys.
- [x] 2.2 Seed valid, invalid, and stale Builder workflow graphs with deterministic draft/version identifiers and expected graph diff data.
- [x] 2.3 Seed running, suspended, and failed runtime instances with tenant-scoped events, safe context, chronological timeline data, sanitized debug traces, and expected audit records.
- [x] 2.4 Restrict browser/service egress to the local test stack, provide deterministic local integration references only, and fail tests on attempted real LLM, RAG, MCP, or provider traffic.
- [x] 2.5 Add fixture validation proving data is synthetic, tenant-isolated, and free of credentials, raw prompts/responses, and RAG documents.

## 3. State Builder golden journey

- [x] 3.1 Implement Editor sign-in and policy-authorized Builder navigation using the golden fixture identity.
- [x] 3.2 Implement the draft edit/save/reload checkpoint and verify the persisted graph through the test verification utility.
- [x] 3.3 Implement valid publish, newest-first version-history, and ordered two-version graph-diff checkpoints.
- [x] 3.4 Implement invalid-graph publish prevention and assert that no publish mutation is sent.
- [x] 3.5 Implement stale-save conflict coverage and assert visible reload guidance plus preservation of the local graph.

## 4. Operator runtime golden journey

- [x] 4.1 Implement Operator sign-in, authorized landing, runtime-list filtering, and tenant-A-only instance discovery checkpoints.
- [x] 4.2 Implement runtime-detail checkpoints for workflow/version, current state, safe context, and chronological timeline entries.
- [x] 4.3 Implement Debug View checkpoints for permitted sanitized provider metadata and the absence of raw provider/RAG data or browser provider requests.
- [x] 4.4 Implement confirmed suspend, resume, and retry journeys against separate state-specific fixture instances, then verify persisted outcomes and audit actor/action records.
- [x] 4.5 Implement direct denied-route and unavailable-action checks for Operator/Viewer, asserting access-denied feedback and no protected data request or cross-tenant data exposure.

## 5. CI, diagnostics, and documentation

- [x] 5.1 Add a dedicated CI browser-E2E job with dependency/browser installation, disposable services, migrations, seeding, server lifecycle, and relevant path triggers.
- [x] 5.2 Capture bounded, synthetic-only failure traces/screenshots/video/log diagnostics and redact or reject disallowed values before upload.
- [x] 5.3 Document local prerequisites, commands, fixture reset behavior, golden checkpoint update rules, and dependency ordering with Phases 1/3/4/5.
- [x] 5.4 Run browser golden journeys repeatedly to prove determinism, then run the existing frontend and backend quality gates to confirm no regression.
- [x] 5.5 Validate the OpenSpec change strictly and resolve all validation issues.
