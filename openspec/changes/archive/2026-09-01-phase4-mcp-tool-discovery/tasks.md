## 1. Catalog persistence

- [x] 1.1 Verify Phase 3 connection lifecycle/API contracts are available and identify the extension points for discovery.
- [x] 1.2 Add migrations for discovery runs and current discovered-tool records with tool identity, schemas, fingerprints, lifecycle, and safe error status.
- [x] 1.3 Add sqlc queries and regenerate database code for scoped catalog reads, reconciliation, and tool enablement updates.
- [x] 1.4 Add domain entities/repository ports for discovery runs and project-scoped tools.

## 2. Discovery service

- [x] 2.1 Extend the MCP transport adapter to initialize a connection and request `tools/list` without any `tools/call` path.
- [x] 2.2 Add provider metadata/schema size limits and redaction/classification for discovery responses and failures.
- [x] 2.3 Implement atomic catalog reconciliation for new, changed, unchanged, and removed tools.
- [x] 2.4 Preserve the prior successful catalog when discovery fails and record a safe failed discovery run.
- [x] 2.5 Implement per-tool enable/disable commands with project authorization and redacted audits.

## 3. APIs and UI integration

- [x] 3.1 Add scoped list-tools, refresh-catalog, and tool-enablement HTTP endpoints with DTO validation.
- [x] 3.2 Document discovery/catalog API behavior and non-execution guarantee in split OpenAPI docs.
- [x] 3.3 Add shared schemas/types/constants/hooks for catalog query and mutations.
- [x] 3.4 Add connection-detail/catalog UI showing tool descriptions, input-schema summary, fingerprints, timestamps, and drift state.
- [x] 3.5 Add explicit refresh and enable/disable controls with retained last-successful results and safe errors.

## 4. Verification

- [x] 4.1 Add backend tests proving discovery calls initialization plus `tools/list` only and never a provider business tool.
- [x] 4.2 Add reconciliation tests for changed schemas, removed tools, failed refreshes, project isolation, and disabled connections/tools.
- [x] 4.3 Add API/UI tests for catalog rendering, refresh feedback, tool enablement, and redaction.
- [x] 4.4 Run migration/sqlc checks, Go tests/vet, web tests/typecheck/lint, provider mock contract tests, and OpenSpec validation.
