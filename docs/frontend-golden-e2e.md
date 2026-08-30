# Frontend golden browser E2E

The browser suite is separate from the web workspace's Vitest/jsdom tests. It
runs Chromium against a real Next.js BFF and Go API, with disposable Postgres
and Redis containers and synthetic tenant data.

## Local prerequisites

- Docker with Compose
- Bun 1.3.5
- Go and the `goose` migration CLI
- Chromium installed with `bun --cwd apps/web exec playwright install chromium`

Run the suite from the repository root:

```bash
go install github.com/pressly/goose/v3/cmd/goose@v3.24.3
bun install
bun --cwd apps/web exec playwright install chromium
bun run test:e2e
```

The runner uses ports `55437` (Postgres), `56381` (Redis), `8021` (API), and
`3020` (web). Override them with `E2E_PG_PORT`, `E2E_REDIS_PORT`,
`E2E_API_PORT`, and `E2E_WEB_PORT` when needed. The default fixture password is
synthetic and is supplied by the runner, not stored in a manifest; set
`E2E_FIXTURE_PASSWORD` to use another local value.

## Fixture reset and verification

Each run starts a fresh Compose project, applies all Goose migrations, truncates
only that disposable database, and runs the internal
`apps/api/cmd/e2e-fixtures` seed command. The command is guarded by
`E2E_FIXTURES=1` and is not an API route. It creates fixed tenant-A Editor,
Operator, Viewer, and sentinel tenant-B identities plus Builder, runtime,
context, event, and audit fixtures. It also fails when disallowed sensitive JSON
fields are present. The Compose project and volumes are removed on exit.

Useful internal checks are available through the root scripts:

```bash
E2E_FIXTURES=1 DATABASE_URL=... E2E_FIXTURE_PASSWORD=... bun run test:e2e:fixtures
E2E_FIXTURES=1 DATABASE_URL=... E2E_FIXTURE_PASSWORD=... bun run test:e2e:fixtures:verify
```

## Golden checkpoint rules

Manifests live in `apps/web/e2e/fixtures/golden-journeys.ts`. Update semantic
identifiers and expected business outcomes there when an intentional product
contract changes. Do not use screenshots, CSS selectors, arbitrary sleeps, raw
provider payloads, credentials, prompts, responses, or RAG documents as golden
data. The fixture safety validator runs before browser execution, and the
network guard aborts non-local browser requests.

Failure screenshots, videos, traces, JSON results, and bounded service logs are
diagnostics only. CI validates their synthetic-data policy before uploading
them, with a 10 MiB per-file and 40 MiB total bound and three-day retention.

## Dependency ordering

The harness is downstream of the Builder lifecycle, Runtime Inspector, Admin
Console, and Auth/RBAC UI changes. Its Builder and Operator journeys exercise
those contracts through the real BFF and API. Update the golden checkpoints
only when an intentional contract change has landed in the corresponding
upstream phase; keep fixture identifiers and safety rules stable across those
updates.
