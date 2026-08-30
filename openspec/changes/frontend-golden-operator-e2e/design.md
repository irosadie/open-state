## Context

The web workspace currently runs Vitest in jsdom and has no browser E2E runner. The backend already has deterministic engine golden conversations and MCP E2E tests, but those do not exercise the Next.js BFF, browser authentication, route/action guards, React Flow persistence, Admin Console controls, or Runtime Inspector presentation.

This change verifies the user-facing contracts introduced by the active Builder lifecycle, Runtime Inspector, Admin Console, and Auth/RBAC UI changes. It is deliberately downstream of those changes and must consume their public routes and APIs rather than duplicate their tests.

## Goals / Non-Goals

### Goals

- Run stable browser journeys against a real Next.js application and API in an isolated test stack.
- Make expected business checkpoints explicit and version controlled as golden fixtures.
- Verify the critical Editor and Operator flows across authentication, authorization, persistence, and UI feedback.
- Provide actionable CI evidence without leaking credentials, prompts, documents, or production data.

### Non-Goals

- Using browser tests to replace unit, component, deterministic-engine, API, or MCP E2E tests.
- Pixel-perfect screenshot comparison or testing real external providers.
- Adding test-only public HTTP endpoints or relaxing production authorization for fixtures.

## Decisions

### 1. Use a dedicated real-browser suite, separate from Vitest

The web workspace adds a browser E2E runner with a distinct `test:e2e` command and configuration. It starts the production-shaped Next.js BFF and API against ephemeral Postgres/Redis, then executes the browser journeys. Unit/component tests remain in Vitest/jsdom and do not need browser/runtime services.

Alternative: test journeys only with React Testing Library. Rejected because it cannot prove browser cookies, proxy forwarding, route redirects, client navigation, and real persistence work together.

### 2. Define golden journeys as semantic checkpoint manifests

Each journey has a checked-in fixture manifest containing synthetic identities, tenant-scoped resource identifiers, input actions, and expected semantic checkpoints. Assertions cover stable labels/statuses, graph/business identifiers, persisted workflow version/state, timeline ordering, and authorization outcomes. Test code uses stable UI affordances and waits for observable responses, never arbitrary timing delays.

Screenshots, video, trace, and request logs are diagnostics on failure only. They do not define correctness, so legitimate visual styling changes do not churn golden baselines.

### 3. Seed and verify via non-production internal test utilities

Before each isolated run, a test-only command creates a fresh synthetic tenant set, an Editor, an Operator, an unauthorized Viewer/sentinel tenant, Builder graphs, runtime instances, events, sanitized traces, and expected audit records. After browser actions, a test verification utility reads persisted state through internal repositories or the disposable database to assert the workflow graph/version, instance lifecycle outcome, tenant isolation, and audit actor/action.

The browser uses only normal product routes and authenticated BFF/API calls. No fixture seeding or verification API is exposed to a production browser.

### 4. Make external integrations impossible in the suite

The test environment uses deterministic local provider stubs only where the platform needs an integration reference. Browser and service egress are restricted to the local test stack. Fixture traces may contain sanitized provider aliases, status, duration, and correlation data, but never credentials, raw prompts/responses, or RAG documents. A request to a non-local LLM, RAG, MCP, or provider endpoint fails the suite.

### 5. Cover two high-value golden stories and explicit negative boundaries

The **Builder golden journey** authenticates an Editor, opens a seeded workflow, makes a deterministic graph edit, observes save/reload persistence, publishes a valid draft, then checks history and a two-version graph diff. It separately exercises invalid publish prevention and stale-save conflict feedback without data loss.

The **Operator golden journey** authenticates an Operator, opens the authorized runtime list/detail, verifies current state/context/timeline and sanitized Debug View, then confirms suspend, resume, and retry against purpose-built fixture instances. It verifies each persisted outcome and audit record. It also verifies that an Operator cannot use unrelated management surfaces and that a denied route does not load protected data.

Alternative: one large story that mutates a single instance through all states. Rejected because failure diagnosis and retry behavior are clearer with fixed, state-specific fixtures.

### 6. Run browser E2E in its own CI job with protected diagnostics

CI installs the browser runtime, prepares disposable services, runs migrations and fixture setup, starts API and web servers, executes the E2E command, and tears down services. On failure, it uploads only artifacts generated from synthetic fixtures; artifact retention is bounded. The job runs on relevant web, API, package, test-infrastructure, and workflow changes.

## Fixture Baseline

| Fixture | Actor | Starting condition | Golden checkpoints |
| --- | --- | --- | --- |
| Builder draft | Editor | Valid saved draft in tenant A | edit, saved status, reload retains graph |
| Builder publish/history | Editor | Valid draft plus two published graph versions | publish version, newest-first history, ordered graph diff |
| Builder conflict/invalid | Editor | Stale draft and invalid graph variants | no invalid publish request; conflict preserves local edit |
| Runtime inspect | Operator | Tenant-A instance with events and sanitized trace | list/detail, current state, chronological timeline, allowed Debug View |
| Runtime commands | Operator | separate running, suspended, and failed instances | confirmed suspend/resume/retry, persisted result, audit actor/action |
| Permission boundary | Operator/Viewer | protected routes in tenant A and sentinel tenant B | authorized landing, direct denial, no protected data request, no cross-tenant data |

## Risks / Trade-offs

- [Browser suites are slower and more operationally complex] → keep two focused journeys, create services once per worker, use deterministic fixtures, and retain unit tests for narrow cases.
- [Shared fixture mutations cause order dependence] → provision fresh database/tenant data per run and separate command instances by starting lifecycle state.
- [UI copy or DOM implementation churn makes tests fragile] → assert semantic contracts through stable labels/affordances and business outcomes, not CSS selectors or full DOM snapshots.
- [Failure artifacts may contain sensitive information] → use synthetic data only, forbid sensitive fixture values, redact logs, and bound artifact retention.
- [Dependent Phase 1/3/4/5 routes land at different times] → implement the harness first, then enable each golden manifest only after its owning contract is present; no local substitute implementation is permitted.

## Migration Plan

1. Add the browser runner, scripts, isolated service orchestration, and empty fixture/verification contract.
2. Seed deterministic Builder and runtime data, then implement the Builder golden journey.
3. Implement the Operator golden journey and negative authorization boundary checks.
4. Add the dedicated CI job, failure diagnostics, and documentation; retain existing Vitest and Go quality gates unchanged.

Rollback removes the separate browser E2E command/job and its test-only fixture tooling. It does not alter production schema, routes, authorization, or persisted application data.
